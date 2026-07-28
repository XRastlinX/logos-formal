package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type HTTPProbeResult struct {
	Schema            string   `json:"schema"`
	Status            string   `json:"status"`
	BaseURL           string   `json:"baseUrl"`
	ValidatedAliases  []string `json:"validatedAliases"`
	GovernanceState   string   `json:"governanceState"`
	AuthorityEffect   string   `json:"authorityEffect"`
	Effect            string   `json:"effect"`
	Forwarded         bool     `json:"forwarded"`
	ApplyProbeStatus  int      `json:"applyProbeStatus"`
	ApplyProbeReason  string   `json:"applyProbeReason"`
	ExternalityStatus string   `json:"externalityStatus"`
	SourceRef         string   `json:"sourceRef"`
	ServerInstance    string   `json:"serverInstance,omitempty"`
}

func requireHTTPStatus(response *http.Response, expected int, label string) error {
	if response.StatusCode == expected {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return fmt.Errorf("%s returned HTTP %d, expected %d: %s",
		label, response.StatusCode, expected, strings.TrimSpace(string(body)))
}

func requireHTTPGovernance(response *http.Response, decision string) error {
	expected := map[string]string{
		"X-Cubed-Bit":        "010",
		"X-Authority-Effect": "NONE",
		"X-Router-Decision":  decision,
		"X-Effect":           "NOT_PERFORMED",
		"X-Forwarded":        "false",
	}
	for name, want := range expected {
		if got := response.Header.Get(name); got != want {
			return fmt.Errorf("%s header %s = %q, expected %q", response.Request.URL.Path, name, got, want)
		}
	}
	return nil
}

func readLimitedBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxHTTPBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxHTTPBodyBytes {
		return nil, errors.New("response body exceeds probe limit")
	}
	return body, nil
}

func runHTTPProbe(baseURL, sourceRef, expectedInstance string) (HTTPProbeResult, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return HTTPProbeResult{}, errors.New("base URL is required")
	}
	sourceRef = strings.TrimSpace(sourceRef)
	if sourceRef == "" {
		sourceRef = detectSourceRef()
	}
	client := &http.Client{Timeout: 5 * time.Second}
	result := HTTPProbeResult{
		Schema:            "urn:cyonic:http-probe:v1",
		Status:            "FAILED",
		BaseURL:           baseURL,
		GovernanceState:   "010",
		AuthorityEffect:   "NONE",
		Effect:            "NOT_PERFORMED",
		Forwarded:         false,
		ExternalityStatus: "NOT_ADJUDICATED",
		SourceRef:         sourceRef,
	}

	health, err := client.Get(baseURL + "/health")
	if err != nil {
		return result, fmt.Errorf("health request: %w", err)
	}
	if err := requireHTTPStatus(health, http.StatusOK, "health"); err != nil {
		health.Body.Close()
		return result, err
	}
	if err := requireHTTPGovernance(health, "OBSERVE_ONLY"); err != nil {
		health.Body.Close()
		return result, err
	}
	healthBody, err := readLimitedBody(health)
	if err != nil {
		return result, fmt.Errorf("read health response: %w", err)
	}
	var healthPayload HTTPResponse
	if err := json.Unmarshal(healthBody, &healthPayload); err != nil {
		return result, fmt.Errorf("decode health response: %w", err)
	}
	result.ServerInstance = healthPayload.InstanceID
	expectedInstance = strings.TrimSpace(expectedInstance)
	if expectedInstance != "" && healthPayload.InstanceID != expectedInstance {
		return result, fmt.Errorf("health instanceId = %q, expected %q",
			healthPayload.InstanceID, expectedInstance)
	}

	sampleResponse, err := client.Get(baseURL + "/api/demo/request")
	if err != nil {
		return result, fmt.Errorf("demo request: %w", err)
	}
	if err := requireHTTPStatus(sampleResponse, http.StatusOK, "demo request"); err != nil {
		sampleResponse.Body.Close()
		return result, err
	}
	if err := requireHTTPGovernance(sampleResponse, "OBSERVE_ONLY"); err != nil {
		sampleResponse.Body.Close()
		return result, err
	}
	sample, err := readLimitedBody(sampleResponse)
	if err != nil {
		return result, fmt.Errorf("read demo request: %w", err)
	}

	aliases := []string{"/api/service/cyonic-validate", "/api/cyonic/validate"}
	for _, path := range aliases {
		request, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(sample))
		if err != nil {
			return result, err
		}
		request.Header.Set("Content-Type", "application/json")
		response, err := client.Do(request)
		if err != nil {
			return result, fmt.Errorf("%s request: %w", path, err)
		}
		if err := requireHTTPStatus(response, http.StatusOK, path); err != nil {
			response.Body.Close()
			return result, err
		}
		if err := requireHTTPGovernance(response, "OBSERVE_ONLY"); err != nil {
			response.Body.Close()
			return result, err
		}
		body, err := readLimitedBody(response)
		if err != nil {
			return result, err
		}
		var payload HTTPResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			return result, fmt.Errorf("decode %s response: %w", path, err)
		}
		if payload.Receipt == nil ||
			payload.Receipt.Effect.Status != "NOT_PERFORMED" ||
			payload.Receipt.Routing.Forwarded ||
			payload.Receipt.Authorization.ServiceAuthorityEffect != "NONE" {
			return result, fmt.Errorf("%s returned an unbounded receipt", path)
		}
		result.ValidatedAliases = append(result.ValidatedAliases, path)
	}

	applyRequest, err := http.NewRequest(http.MethodPost, baseURL+"/api/service/apply",
		strings.NewReader(`{"permitMaterial":"MUST_NOT_BE_READ"}`))
	if err != nil {
		return result, err
	}
	applyRequest.Header.Set("Content-Type", "application/json")
	applyResponse, err := client.Do(applyRequest)
	if err != nil {
		return result, fmt.Errorf("Apply compatibility probe: %w", err)
	}
	result.ApplyProbeStatus = applyResponse.StatusCode
	if err := requireHTTPStatus(applyResponse, http.StatusMethodNotAllowed, "Apply compatibility probe"); err != nil {
		applyResponse.Body.Close()
		return result, err
	}
	if err := requireHTTPGovernance(applyResponse, "REJECT"); err != nil {
		applyResponse.Body.Close()
		return result, err
	}
	applyBody, err := readLimitedBody(applyResponse)
	if err != nil {
		return result, err
	}
	var applyPayload HTTPResponse
	if err := json.Unmarshal(applyBody, &applyPayload); err != nil {
		return result, fmt.Errorf("decode Apply rejection: %w", err)
	}
	result.ApplyProbeReason = applyPayload.ReasonCode
	if applyPayload.ReasonCode != "EFFECT_ROUTE_FORBIDDEN" ||
		applyPayload.AuthorityEffect != "NONE" ||
		applyPayload.Effect != "NOT_PERFORMED" ||
		applyPayload.Forwarded {
		return result, errors.New("Apply compatibility path did not fail closed")
	}

	result.Status = "COLD_CALL_PASSED"
	return result, nil
}

func smokeHTTP(now time.Time) (HTTPProbeResult, error) {
	fixture, err := makeFixture(now)
	if err != nil {
		return HTTPProbeResult{}, err
	}
	config := HTTPConfig{
		TrustedIssuer: "example-principal",
		PublicKey:     fixture.PublicKey,
		OriginClaim:   "undetermined",
		InstanceID:    "smoke-http",
		Now:           func() time.Time { return now },
		SampleRequest: &fixture.Request,
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return HTTPProbeResult{}, fmt.Errorf("open loopback listener: %w", err)
	}
	server := &http.Server{
		Handler:           newHTTPHandler(config),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	result, probeErr := runHTTPProbe(
		"http://"+listener.Addr().String(),
		detectSourceRef(),
		config.InstanceID,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	shutdownErr := server.Shutdown(ctx)
	cancel()
	serveErr := <-serverErrors
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return result, fmt.Errorf("serve smoke surface: %w", serveErr)
	}
	if shutdownErr != nil {
		return result, fmt.Errorf("stop smoke surface: %w", shutdownErr)
	}
	if probeErr != nil {
		return result, probeErr
	}
	return result, nil
}

func runSmokeHTTP(args []string) int {
	flags := flag.NewFlagSet("smoke-http", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "smoke-http: no positional arguments are accepted")
		return 2
	}
	result, err := smokeHTTP(time.Now().UTC())
	if err != nil {
		fmt.Fprintln(os.Stderr, "smoke-http:", err)
		return 2
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, "smoke-http:", err)
		return 2
	}
	return 0
}

func runProbeHTTP(args []string) int {
	flags := flag.NewFlagSet("probe-http", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	baseURL := flags.String("base-url", "http://127.0.0.1:8787", "Cyonic service base URL")
	sourceRef := flags.String("source-ref", "", "explicit source commit or artifact reference")
	expectedInstance := flags.String("expected-instance", "", "required server instance identity when process binding is needed")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	result, err := runHTTPProbe(*baseURL, *sourceRef, *expectedInstance)
	if err != nil {
		result.Status = "COLD_CALL_FAILED"
		_ = json.NewEncoder(os.Stdout).Encode(result)
		fmt.Fprintln(os.Stderr, "probe-http:", err)
		return 2
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, "probe-http:", err)
		return 2
	}
	return 0
}
