// Status: PROPOSED
// authority_effect: NONE
// Package transport provides HTTP loopback stubs for V0.2 demonstrator.
package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
)

// LoopbackClient sends envelopes to a cyxd daemon via HTTP stub.
type LoopbackClient struct {
	BaseURL string
}

func NewLoopbackClient(url string) *LoopbackClient {
	return &LoopbackClient{
		BaseURL: url,
	}
}

func (c *LoopbackClient) SendEnvelope(env *envelope.CyExchangeEnvelope) error {
	payload, err := json.Marshal(env)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/inbox", c.BaseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("transport failure: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("node rejected envelope, status %d", resp.StatusCode)
	}
	return nil
}
