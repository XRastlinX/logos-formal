package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testHTTPConfig(t *testing.T) (HTTPConfig, BoundaryRequest) {
	t.Helper()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	fixture, err := makeFixture(now)
	if err != nil {
		t.Fatal(err)
	}
	return HTTPConfig{
		TrustedIssuer: "example-principal",
		PublicKey:     fixture.PublicKey,
		OriginClaim:   "undetermined",
		Now:           func() time.Time { return now },
		SampleRequest: &fixture.Request,
	}, fixture.Request
}

func assertGovernanceHeaders(t *testing.T, response *http.Response, decision string) {
	t.Helper()
	expected := map[string]string{
		"X-Cubed-Bit":        "010",
		"X-Authority-Effect": "NONE",
		"X-Router-Decision":  decision,
		"X-Effect":           "NOT_PERFORMED",
		"X-Forwarded":        "false",
	}
	for name, want := range expected {
		if got := response.Header.Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func postBoundaryRequest(t *testing.T, server *httptest.Server, path string, request BoundaryRequest) *http.Response {
	t.Helper()
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	httpRequest, err := http.NewRequest(http.MethodPost, server.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	response, err := server.Client().Do(httpRequest)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func decodeHTTPResponse(t *testing.T, response *http.Response) HTTPResponse {
	t.Helper()
	defer response.Body.Close()
	var payload HTTPResponse
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestHTTPValidateAliasesAreBoundedAndEquivalent(t *testing.T) {
	config, request := testHTTPConfig(t)
	server := httptest.NewServer(newHTTPHandler(config))
	defer server.Close()

	for _, path := range []string{"/api/service/cyonic-validate", "/api/cyonic/validate"} {
		t.Run(path, func(t *testing.T) {
			response := postBoundaryRequest(t, server, path, request)
			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}
			assertGovernanceHeaders(t, response, "OBSERVE_ONLY")
			payload := decodeHTTPResponse(t, response)
			if payload.AuthorityEffect != "NONE" || payload.Effect != "NOT_PERFORMED" || payload.Forwarded {
				t.Fatalf("unbounded response: %+v", payload)
			}
			if payload.Receipt == nil {
				t.Fatal("receipt is missing")
			}
			if payload.Receipt.Routing.Decision != "OBSERVE_ONLY" ||
				payload.Receipt.Routing.Forwarded ||
				payload.Receipt.Effect.Status != "NOT_PERFORMED" {
				t.Fatalf("unbounded receipt: %+v", payload.Receipt)
			}
		})
	}
}

func TestHTTPRejectsEffectfulOperation(t *testing.T) {
	config, request := testHTTPConfig(t)
	request.Operation = "APPLY"
	server := httptest.NewServer(newHTTPHandler(config))
	defer server.Close()

	response := postBoundaryRequest(t, server, "/api/service/cyonic-validate", request)
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
	assertGovernanceHeaders(t, response, "REJECT")
	payload := decodeHTTPResponse(t, response)
	if payload.ReasonCode != "EFFECT_ROUTE_FORBIDDEN" {
		t.Fatalf("reasonCode = %q", payload.ReasonCode)
	}
	if payload.Receipt == nil || payload.Receipt.Routing.Forwarded ||
		payload.Receipt.Effect.Status != "NOT_PERFORMED" {
		t.Fatalf("effectful rejection crossed boundary: %+v", payload)
	}
}

type countedReader struct {
	reads int
}

func (reader *countedReader) Read(_ []byte) (int, error) {
	reader.reads++
	return 0, io.EOF
}

func TestHTTPApplyRejectsBeforeReadingBody(t *testing.T) {
	config, _ := testHTTPConfig(t)
	handler := newHTTPHandler(config)
	body := &countedReader{}
	request := httptest.NewRequest(http.MethodPost, "/api/service/apply", body)
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)
	response := responseRecorder.Result()
	defer response.Body.Close()

	if body.reads != 0 {
		t.Fatalf("Apply route read request body %d time(s)", body.reads)
	}
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusMethodNotAllowed)
	}
	assertGovernanceHeaders(t, response, "REJECT")
	var payload HTTPResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.ReasonCode != "EFFECT_ROUTE_FORBIDDEN" ||
		payload.AuthorityEffect != "NONE" ||
		payload.Effect != "NOT_PERFORMED" ||
		payload.Forwarded {
		t.Fatalf("Apply rejection is not fail-closed: %+v", payload)
	}
}

func TestHTTPGovernanceHeadersCoverHealthErrorsAndUnknownRoutes(t *testing.T) {
	config, _ := testHTTPConfig(t)
	server := httptest.NewServer(newHTTPHandler(config))
	defer server.Close()

	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
		decision string
	}{
		{"health", http.MethodGet, "/health", http.StatusOK, "OBSERVE_ONLY"},
		{"health method", http.MethodPost, "/health", http.StatusMethodNotAllowed, "REJECT"},
		{"unknown", http.MethodGet, "/api/no-such-route", http.StatusNotFound, "REJECT"},
		{"validate method", http.MethodGet, "/api/cyonic/validate", http.StatusMethodNotAllowed, "REJECT"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(test.method, server.URL+test.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := server.Client().Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.wantCode {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.wantCode)
			}
			assertGovernanceHeaders(t, response, test.decision)
		})
	}
}

func TestHTTPRejectsMalformedOversizedAndWrongMediaType(t *testing.T) {
	config, _ := testHTTPConfig(t)
	server := httptest.NewServer(newHTTPHandler(config))
	defer server.Close()

	tests := []struct {
		name        string
		contentType string
		body        string
		contentLen  int64
		wantCode    int
		wantReason  string
	}{
		{"wrong media", "text/plain", "{}", 2, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE"},
		{"malformed", "application/json", "{", 1, http.StatusBadRequest, "MALFORMED_REQUEST"},
		{"oversized", "application/json", strings.Repeat(" ", int(maxHTTPBodyBytes+1)), maxHTTPBodyBytes + 1, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost,
				server.URL+"/api/service/cyonic-validate", strings.NewReader(test.body))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Content-Type", test.contentType)
			request.ContentLength = test.contentLen
			response, err := server.Client().Do(request)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != test.wantCode {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.wantCode)
			}
			assertGovernanceHeaders(t, response, "REJECT")
			payload := decodeHTTPResponse(t, response)
			if payload.ReasonCode != test.wantReason {
				t.Fatalf("reasonCode = %q, want %q", payload.ReasonCode, test.wantReason)
			}
		})
	}
}

func TestHTTPDemoRequestIsExplicitlyNonAuthoritative(t *testing.T) {
	config, _ := testHTTPConfig(t)
	server := httptest.NewServer(newHTTPHandler(config))
	defer server.Close()

	response, err := server.Client().Get(server.URL + "/api/demo/request")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	assertGovernanceHeaders(t, response, "OBSERVE_ONLY")
	if got := response.Header.Get("X-Cyonic-Sample"); got != "SIGNED_FIXTURE_ONLY" {
		t.Fatalf("X-Cyonic-Sample = %q", got)
	}
	defer response.Body.Close()
	var sample BoundaryRequest
	if err := json.NewDecoder(response.Body).Decode(&sample); err != nil {
		t.Fatal(err)
	}
	if sample.Operation != "VERIFY_PERMIT_EVIDENCE" {
		t.Fatalf("demo request operation = %q", sample.Operation)
	}
}

func TestHTTPProbeExercisesAliasesAndFailClosedApply(t *testing.T) {
	config, _ := testHTTPConfig(t)
	server := httptest.NewServer(newHTTPHandler(config))
	defer server.Close()

	result, err := runHTTPProbe(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "COLD_CALL_PASSED" {
		t.Fatalf("status = %q", result.Status)
	}
	if len(result.ValidatedAliases) != 2 {
		t.Fatalf("validated aliases = %v", result.ValidatedAliases)
	}
	if result.ApplyProbeStatus != http.StatusMethodNotAllowed ||
		result.ApplyProbeReason != "EFFECT_ROUTE_FORBIDDEN" {
		t.Fatalf("Apply probe did not fail closed: %+v", result)
	}
	if result.AuthorityEffect != "NONE" || result.Effect != "NOT_PERFORMED" || result.Forwarded {
		t.Fatalf("probe implied authority or effect: %+v", result)
	}
	if result.ExternalityStatus != "NOT_ADJUDICATED" {
		t.Fatalf("probe self-adjudicated externality: %+v", result)
	}
}

func TestHTTPSmokeRunsOneCommandBoundaryWithoutExternalityClaim(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 30, 0, 0, time.UTC)
	result, err := smokeHTTP(now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "COLD_CALL_PASSED" ||
		result.GovernanceState != "010" ||
		result.AuthorityEffect != "NONE" ||
		result.Effect != "NOT_PERFORMED" ||
		result.Forwarded ||
		result.ApplyProbeStatus != http.StatusMethodNotAllowed ||
		result.ApplyProbeReason != "EFFECT_ROUTE_FORBIDDEN" {
		t.Fatalf("smoke result crossed boundary: %+v", result)
	}
	if result.ExternalityStatus != "NOT_ADJUDICATED" {
		t.Fatalf("smoke result self-adjudicated externality: %+v", result)
	}
}
