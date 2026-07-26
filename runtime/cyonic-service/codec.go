package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	artifactDomain = "cyonic-artifact-v1"
	permitDomain   = "cyonic-permit-v1"
)

func appendField(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.BigEndian, uint32(len(value)))
	buf.WriteString(value)
}

func artifactBytes(artifact Artifact) []byte {
	var buf bytes.Buffer
	appendField(&buf, artifactDomain)
	appendField(&buf, artifact.Proposer)
	appendField(&buf, artifact.Action)
	appendField(&buf, artifact.Target)
	appendField(&buf, artifact.PayloadDigest)
	return buf.Bytes()
}

func hashArtifact(artifact Artifact) string {
	digest := sha256.Sum256(artifactBytes(artifact))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func permitSigningBytes(permit Permit) []byte {
	var buf bytes.Buffer
	appendField(&buf, permitDomain)
	appendField(&buf, permit.Schema)
	appendField(&buf, permit.ArtifactDigest)
	appendField(&buf, permit.Action)
	appendField(&buf, permit.Target)
	appendField(&buf, permit.Issuer)
	appendField(&buf, permit.ExpiresAt)
	appendField(&buf, permit.Nonce)
	return buf.Bytes()
}

func signPermit(permit *Permit, privateKey ed25519.PrivateKey) {
	permit.Signature = base64.StdEncoding.EncodeToString(
		ed25519.Sign(privateKey, permitSigningBytes(*permit)),
	)
}

func verifyPermitSignature(permit Permit, publicKey ed25519.PublicKey) error {
	signature, err := base64.StdEncoding.DecodeString(permit.Signature)
	if err != nil {
		return fmt.Errorf("decode permit signature: %w", err)
	}
	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("permit signature has %d bytes, expected %d", len(signature), ed25519.SignatureSize)
	}
	if !ed25519.Verify(publicKey, permitSigningBytes(permit), signature) {
		return errors.New("Ed25519 signature verification failed")
	}
	return nil
}

func encodePublicKey(publicKey ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(publicKey)
}

func loadPublicKey(path string) (ed25519.PublicKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read trust key: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("decode trust key: %w", err)
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("trust key has %d bytes, expected %d", len(decoded), ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(decoded), nil
}

func validSHA256Digest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}
