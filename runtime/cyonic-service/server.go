package main

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultListenAddress = "127.0.0.1:8787"
	maxHTTPBodyBytes     = int64(1 << 20)
	httpSchema           = "urn:cyonic:http-response:v1"
)

type HTTPConfig struct {
	TrustedIssuer string
	PublicKey     ed25519.PublicKey
	OriginClaim   string
	InstanceID    string
	Now           func() time.Time
	SampleRequest *BoundaryRequest
}

type HTTPResponse struct {
	Schema          string           `json:"schema"`
	Status          string           `json:"status"`
	GovernanceState string           `json:"governanceState"`
	AuthorityEffect string           `json:"authorityEffect"`
	Effect          string           `json:"effect"`
	Forwarded       bool             `json:"forwarded"`
	InstanceID      string           `json:"instanceId,omitempty"`
	ReasonCode      string           `json:"reasonCode,omitempty"`
	Detail          string           `json:"detail"`
	Receipt         *BoundaryReceipt `json:"receipt,omitempty"`
	SampleRequest   *BoundaryRequest `json:"sampleRequest,omitempty"`
}

func setGovernanceHeaders(header http.Header, decision string) {
	header.Set("Cache-Control", "no-store")
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Cubed-Bit", "010")
	header.Set("X-Authority-Effect", "NONE")
	header.Set("X-Router-Decision", decision)
	header.Set("X-Effect", "NOT_PERFORMED")
	header.Set("X-Forwarded", "false")
}

func writeHTTPJSON(writer http.ResponseWriter, statusCode int, decision string, response HTTPResponse) {
	setGovernanceHeaders(writer.Header(), decision)
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(response)
}

func baseHTTPResponse(status, detail string) HTTPResponse {
	return HTTPResponse{
		Schema:          httpSchema,
		Status:          status,
		GovernanceState: "010",
		AuthorityEffect: "NONE",
		Effect:          "NOT_PERFORMED",
		Forwarded:       false,
		Detail:          detail,
	}
}

func requestContentTypeIsJSON(request *http.Request) bool {
	contentType := request.Header.Get("Content-Type")
	if contentType == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}

func decodeHTTPBoundaryRequest(writer http.ResponseWriter, request *http.Request) (BoundaryRequest, error) {
	if request.ContentLength > maxHTTPBodyBytes {
		return BoundaryRequest{}, fmt.Errorf("request body exceeds %d bytes", maxHTTPBodyBytes)
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxHTTPBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	var boundaryRequest BoundaryRequest
	if err := decoder.Decode(&boundaryRequest); err != nil {
		return BoundaryRequest{}, fmt.Errorf("decode request: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return BoundaryRequest{}, err
	}
	return boundaryRequest, nil
}

func statusForEvaluationCode(code int) int {
	switch code {
	case 0:
		return http.StatusOK
	case 1:
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}

func newHTTPHandler(config HTTPConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			response := baseHTTPResponse("REJECTED", "Health is available through GET only.")
			response.ReasonCode = "METHOD_NOT_ALLOWED"
			writer.Header().Set("Allow", http.MethodGet)
			writeHTTPJSON(writer, http.StatusMethodNotAllowed, "REJECT", response)
			return
		}
		response := baseHTTPResponse(
			"READY",
			"Cyonic validation is available; this process has no Apply or forwarding capability.",
		)
		response.InstanceID = config.InstanceID
		writeHTTPJSON(writer, http.StatusOK, "OBSERVE_ONLY", response)
	})

	validate := func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			response := baseHTTPResponse("REJECTED", "Validation is available through POST only.")
			response.ReasonCode = "METHOD_NOT_ALLOWED"
			writer.Header().Set("Allow", http.MethodPost)
			writeHTTPJSON(writer, http.StatusMethodNotAllowed, "REJECT", response)
			return
		}
		if !requestContentTypeIsJSON(request) {
			response := baseHTTPResponse("REJECTED", "Content-Type must be application/json.")
			response.ReasonCode = "UNSUPPORTED_MEDIA_TYPE"
			writeHTTPJSON(writer, http.StatusUnsupportedMediaType, "REJECT", response)
			return
		}

		boundaryRequest, err := decodeHTTPBoundaryRequest(writer, request)
		if err != nil {
			response := baseHTTPResponse("REJECTED", err.Error())
			response.ReasonCode = "MALFORMED_REQUEST"
			statusCode := http.StatusBadRequest
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) || request.ContentLength > maxHTTPBodyBytes {
				statusCode = http.StatusRequestEntityTooLarge
				response.ReasonCode = "REQUEST_TOO_LARGE"
			}
			writeHTTPJSON(writer, statusCode, "REJECT", response)
			return
		}

		receipt, code := evaluate(boundaryRequest, EvaluationConfig{
			TrustedIssuer: config.TrustedIssuer,
			PublicKey:     config.PublicKey,
			Now:           config.Now,
			OriginClaim:   config.OriginClaim,
		})
		response := baseHTTPResponse(receipt.Verdict,
			"Interpretation and external permit evidence were evaluated; no effect was performed or forwarded.")
		response.Receipt = &receipt
		if code != 0 {
			response.Status = "REJECTED"
			response.ReasonCode = receipt.Friction.Code
			writeHTTPJSON(writer, statusForEvaluationCode(code), "REJECT", response)
			return
		}
		writeHTTPJSON(writer, http.StatusOK, "OBSERVE_ONLY", response)
	}

	mux.HandleFunc("/api/service/cyonic-validate", validate)
	mux.HandleFunc("/api/cyonic/validate", validate)

	mux.HandleFunc("/api/service/apply", func(writer http.ResponseWriter, request *http.Request) {
		response := baseHTTPResponse("REJECTED",
			"This service has no Apply dispatcher. The request body and any credential or permit material were not parsed.")
		response.ReasonCode = "EFFECT_ROUTE_FORBIDDEN"
		writer.Header().Set("Allow", "")
		writeHTTPJSON(writer, http.StatusMethodNotAllowed, "REJECT", response)
	})

	mux.HandleFunc("/api/demo/request", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			response := baseHTTPResponse("REJECTED", "The demonstration request is available through GET only.")
			response.ReasonCode = "METHOD_NOT_ALLOWED"
			writer.Header().Set("Allow", http.MethodGet)
			writeHTTPJSON(writer, http.StatusMethodNotAllowed, "REJECT", response)
			return
		}
		if config.SampleRequest == nil {
			response := baseHTTPResponse("REJECTED", "No demonstration request is configured for this server.")
			response.ReasonCode = "DEMO_REQUEST_UNAVAILABLE"
			writeHTTPJSON(writer, http.StatusNotFound, "REJECT", response)
			return
		}
		setGovernanceHeaders(writer.Header(), "OBSERVE_ONLY")
		writer.Header().Set("X-Cyonic-Sample", "SIGNED_FIXTURE_ONLY")
		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(config.SampleRequest)
	})

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		knownPath := request.URL.Path == "/health" ||
			request.URL.Path == "/api/service/cyonic-validate" ||
			request.URL.Path == "/api/cyonic/validate" ||
			request.URL.Path == "/api/service/apply" ||
			request.URL.Path == "/api/demo/request"
		if !knownPath {
			response := baseHTTPResponse("REJECTED", "No bounded route exists for this path.")
			response.ReasonCode = "ROUTE_NOT_FOUND"
			writeHTTPJSON(writer, http.StatusNotFound, "REJECT", response)
			return
		}
		mux.ServeHTTP(writer, request)
	})
}

func loadHTTPConfig(trustKeyPath, issuer, originClaim, instanceID string) (HTTPConfig, error) {
	if strings.TrimSpace(trustKeyPath) == "" {
		return HTTPConfig{}, errors.New("-trust-key is required")
	}
	if strings.TrimSpace(issuer) == "" {
		return HTTPConfig{}, errors.New("-issuer is required")
	}
	publicKey, err := loadPublicKey(trustKeyPath)
	if err != nil {
		return HTTPConfig{}, err
	}
	return HTTPConfig{
		TrustedIssuer: issuer,
		PublicKey:     publicKey,
		OriginClaim:   originClaim,
		InstanceID:    strings.TrimSpace(instanceID),
	}, nil
}

func serveHTTP(address string, config HTTPConfig) error {
	server := &http.Server{
		Addr:              address,
		Handler:           newHTTPHandler(config),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Printf("Cyonic Service Surface listening on http://%s", address)
	log.Printf("governanceState=010 authorityEffect=NONE effect=NOT_PERFORMED forwarded=false")
	return server.ListenAndServe()
}

func runServe(args []string) int {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	address := flags.String("listen", defaultListenAddress, "listen address")
	trustKeyPath := flags.String("trust-key", "", "trusted Ed25519 public key path")
	issuer := flags.String("issuer", "", "trusted issuer identifier")
	originClaim := flags.String("origin-claim", "undetermined",
		"caller relationship claim recorded in receipts; never self-verifies externality")
	instanceID := flags.String("instance-id", "", "optional caller-selected process identity for bounded health checks")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	config, err := loadHTTPConfig(*trustKeyPath, *issuer, *originClaim, *instanceID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "serve:", err)
		return 2
	}
	if err := serveHTTP(*address, config); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintln(os.Stderr, "serve:", err)
		return 2
	}
	return 0
}

func runServeDemo(args []string) int {
	flags := flag.NewFlagSet("serve-demo", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	address := flags.String("listen", defaultListenAddress, "listen address")
	instanceID := flags.String("instance-id", "", "optional caller-selected process identity for bounded health checks")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	now := time.Now().UTC()
	fixture, err := makeFixture(now)
	if err != nil {
		fmt.Fprintln(os.Stderr, "serve-demo:", err)
		return 2
	}
	config := HTTPConfig{
		TrustedIssuer: "example-principal",
		PublicKey:     fixture.PublicKey,
		OriginClaim:   "undetermined",
		InstanceID:    strings.TrimSpace(*instanceID),
		Now:           func() time.Time { return now },
		SampleRequest: &fixture.Request,
	}
	fmt.Fprintf(os.Stderr, "Demo request: GET http://%s/api/demo/request\n", *address)
	fmt.Fprintf(os.Stderr, "Validate aliases: POST http://%s/api/service/cyonic-validate or /api/cyonic/validate\n", *address)
	fmt.Fprintln(os.Stderr, "The ephemeral private key has been discarded. No Apply route is available.")
	if err := serveHTTP(*address, config); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintln(os.Stderr, "serve-demo:", err)
		return 2
	}
	return 0
}
