// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/discovery"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/storage"
)

// ServeCommand starts the inbox listener.
func ServeCommand(port string) {
	log.Printf("[PROPOSED] Starting CyExchange durable inbox loop on port %s", port)

	dbPath := os.Getenv("CYEXCHANGE_DB_PATH")
	if dbPath == "" {
		dbPath = "cyexchange.db"
	}
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Storage initialization failed: %v", err)
	}
	defer store.Close()

	inboxAdapter, err := recoverInbox(store)
	if err != nil {
		log.Fatalf("Recovery failed: %v", err)
	}

	// Mock NodeCard for the advertiser
	card := &envelope.NodeCard{DID: "did:key:stub"}
	adv := discovery.NewAdvertiser(card, port)
	adv.Broadcast()
	defer adv.Stop()

	prober := discovery.NewProber()
	mux := newHTTPHandler(store, inboxAdapter, prober)
	prober.MarkReady()

	addr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func newHTTPHandler(
	store storage.Store,
	inboxAdapter *inbox.Adapter,
	prober *discovery.Prober,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", prober.Handler())
	mux.HandleFunc("/v1/inbox", func(w http.ResponseWriter, r *http.Request) {
		inboxHandler(store, inboxAdapter, w, r)
	})
	return mux
}

func inboxHandler(
	store storage.Store,
	inboxAdapter *inbox.Adapter,
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	env, err := envelope.DecodeStrict(r.Body, envelope.MaxEnvelopeBytes)
	if err != nil {
		http.Error(w, "Malformed JSON", http.StatusBadRequest)
		return
	}

	if err := envelope.Validate(env); err != nil {
		log.Printf("[REJECT] Envelope validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := trackLineage(env, store); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := enforceReplayImmunity(env, store); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	ttl := 1 * time.Hour
	expiresAt := time.Now().UTC().Add(ttl)
	err = store.InsertEnvelope(env, expiresAt)
	if err != nil {
		log.Printf("[REJECT] Durable storage ingestion failed: %v", err)
		status := http.StatusInternalServerError
		if errors.Is(err, cyerrors.ErrReplay) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}

	err = inboxAdapter.IngestUntil(env, expiresAt)
	if err != nil {
		log.Printf("[REJECT] State machine ingestion failed: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	log.Printf("[RECEIVED] Envelope %s ingested safely. Outputting DeliveryReceipt stub.", env.EnvelopeID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"STATUS_RECEIVED"}`))
}
