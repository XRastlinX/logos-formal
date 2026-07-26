// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/discovery"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
)

// ServeCommand starts the inbox listener.
func ServeCommand(port string) {
	log.Printf("[PROPOSED] Starting CyExchange durable inbox loop on port %s", port)
	
	// Mock NodeCard for the advertiser
	card := &envelope.NodeCard{DID: "did:key:stub"}
	adv := discovery.NewAdvertiser(card, port)
	adv.Broadcast()
	defer adv.Stop()

	http.HandleFunc("/v1/inbox", inboxHandler)

	addr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func inboxHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var env envelope.CyExchangeEnvelope
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		http.Error(w, "Malformed JSON", http.StatusBadRequest)
		return
	}

	err := envelope.Validate(&env)
	if err != nil {
		log.Printf("[REJECT] Envelope validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1-hour TTL hardcoded for PROPOSED stub
	err = globalInbox.Ingest(&env, 1*time.Hour)
	if err != nil {
		log.Printf("[REJECT] State machine ingestion failed: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	log.Printf("[RECEIVED] Envelope %s ingested safely. Outputting DeliveryReceipt stub.", env.EnvelopeID)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"STATUS_RECEIVED"}`))
}
