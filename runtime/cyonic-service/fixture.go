package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type generatedFixture struct {
	Request   BoundaryRequest
	PublicKey ed25519.PublicKey
}

func makeFixture(now time.Time) (generatedFixture, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return generatedFixture{}, fmt.Errorf("generate Ed25519 key: %w", err)
	}

	payload := sha256.Sum256([]byte("hello from an external client"))
	artifact := Artifact{
		Proposer:      "external-client-example",
		Action:        "demo.echo",
		Target:        "example-output",
		PayloadDigest: "sha256:" + hex.EncodeToString(payload[:]),
	}
	permit := Permit{
		Schema:         permitSchema,
		ArtifactDigest: hashArtifact(artifact),
		Action:         artifact.Action,
		Target:         artifact.Target,
		Issuer:         "example-principal",
		ExpiresAt:      now.UTC().Add(15 * time.Minute).Format(time.RFC3339),
		Nonce:          "example-" + now.UTC().Format("20060102T150405.000000000Z"),
	}
	signPermit(&permit, privateKey)

	return generatedFixture{
		PublicKey: publicKey,
		Request: BoundaryRequest{
			Schema:    requestSchema,
			RequestID: "example-request",
			Operation: "VERIFY_PERMIT_EVIDENCE",
			Artifact:  artifact,
			Permit:    permit,
		},
	}, nil
}

func writeFixture(dir string, fixture generatedFixture) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create fixture directory: %w", err)
	}

	requestBytes, err := json.MarshalIndent(fixture.Request, "", "  ")
	if err != nil {
		return fmt.Errorf("encode request fixture: %w", err)
	}
	requestBytes = append(requestBytes, '\n')

	if err := os.WriteFile(filepath.Join(dir, "trusted-public.key"), []byte(encodePublicKey(fixture.PublicKey)+"\n"), 0o644); err != nil {
		return fmt.Errorf("write public key: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "request.json"), requestBytes, 0o644); err != nil {
		return fmt.Errorf("write request fixture: %w", err)
	}
	return nil
}
