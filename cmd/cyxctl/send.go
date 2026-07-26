// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"fmt"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/transport"
)

func SendCommand(args []string) {
	fmt.Println("[PROPOSED] cyxctl send initiated.")
	if len(args) < 2 || args[0] != "--dest" {
		fmt.Println("Usage: cyxctl send --dest <url> --payload <file.json>")
		return
	}

	dest := args[1]
	
	// Create a dummy PROPOSED envelope.
	// In reality, this would load and parse the --payload argument.
	env := &envelope.CyExchangeEnvelope{
		EnvelopeID:    fmt.Sprintf("env_%d", time.Now().UnixNano()),
		Sender:        envelope.NodeCard{DID: "did:key:stub"},
		RequiredState: "010",
		PayloadHash:   "stubhash",
	}

	fmt.Printf("[PROPOSED] Dispatching envelope %s to %s\n", env.EnvelopeID, dest)

	client := transport.NewLoopbackClient(fmt.Sprintf("http://%s", dest))
	err := client.SendEnvelope(env)
	if err != nil {
		fmt.Printf("[ERROR] Dispatch failed: %v\n", err)
		return
	}

	fmt.Println("[PROPOSED] Dispatch succeeded. Envelope accepted by remote inbox.")
}
