// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/discovery"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/storage"
)

func signedHandlerEnvelope(t *testing.T, id, state string) *envelope.CyExchangeEnvelope {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"request":"observe"}`)
	payloadHash, err := envelope.HashPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	env := &envelope.CyExchangeEnvelope{
		EnvelopeID: id,
		Sender: envelope.NodeCard{
			DID:              "did:key:sender",
			PublicKeyEd25519: hex.EncodeToString(publicKey),
			CapabilityScope:  "010-observer",
		},
		Receiver:         envelope.NodeCard{DID: "did:key:receiver"},
		RequiredState:    state,
		PayloadSchemaURI: "urn:cyexchange:test",
		PayloadBytes:     payload,
		PayloadHash:      payloadHash,
	}
	signingBytes, err := envelope.SigningBytes(env)
	if err != nil {
		t.Fatal(err)
	}
	env.SignatureEd25519 = hex.EncodeToString(ed25519.Sign(privateKey, signingBytes))
	return env
}

func TestInboxHandlerAcceptsOnceAndRejectsReplay(t *testing.T) {
	store := storage.NewMemoryStore()
	inboxAdapter := inbox.NewAdapter()
	prober := discovery.NewProber()
	prober.MarkReady()
	handler := newHTTPHandler(store, inboxAdapter, prober)
	body, err := json.Marshal(signedHandlerEnvelope(t, "http-valid", "010"))
	if err != nil {
		t.Fatal(err)
	}

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/v1/inbox", bytes.NewReader(body)))
	if first.Code != http.StatusAccepted {
		t.Fatalf("first response = %d: %s", first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/v1/inbox", bytes.NewReader(body)))
	if second.Code != http.StatusConflict {
		t.Fatalf("replay response = %d: %s", second.Code, second.Body.String())
	}
}

func TestInboxHandlerRejects101AndDuplicateJSONKeys(t *testing.T) {
	store := storage.NewMemoryStore()
	handler := newHTTPHandler(store, inbox.NewAdapter(), discovery.NewProber())

	body, err := json.Marshal(signedHandlerEnvelope(t, "http-101", "101"))
	if err != nil {
		t.Fatal(err)
	}
	actuator := httptest.NewRecorder()
	handler.ServeHTTP(
		actuator,
		httptest.NewRequest(http.MethodPost, "/v1/inbox", bytes.NewReader(body)),
	)
	if actuator.Code != http.StatusBadRequest {
		t.Fatalf("101 response = %d: %s", actuator.Code, actuator.Body.String())
	}

	duplicate := httptest.NewRecorder()
	handler.ServeHTTP(
		duplicate,
		httptest.NewRequest(
			http.MethodPost,
			"/v1/inbox",
			bytes.NewBufferString(`{"envelope_id":"a","envelope_id":"b"}`),
		),
	)
	if duplicate.Code != http.StatusBadRequest {
		t.Fatalf("duplicate-key response = %d: %s", duplicate.Code, duplicate.Body.String())
	}
}
