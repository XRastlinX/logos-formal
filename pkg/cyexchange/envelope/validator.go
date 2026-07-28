// Status: PROPOSED
// authority_effect: NONE
// Package envelope defines strict CyExchange envelope integrity validation.
package envelope

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/crypto"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
)

const MaxEnvelopeBytes = 1024 * 1024

var (
	envelopeIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)
	hashPattern       = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type CyExchangeEnvelope struct {
	EnvelopeID       string   `json:"envelope_id"`
	Sender           NodeCard `json:"sender"`
	Receiver         NodeCard `json:"receiver"`
	RequiredState    string   `json:"required_state"`
	PayloadSchemaURI string   `json:"payload_schema_uri"`
	PayloadBytes     []byte   `json:"payload_bytes"`
	PayloadHash      string   `json:"payload_hash"`
	Lineage          Lineage  `json:"lineage"`
	SignatureEd25519 string   `json:"signature_ed25519,omitempty"`
}

type NodeCard struct {
	DID              string `json:"did"`
	PublicKeyEd25519 string `json:"public_key_ed25519"`
	CapabilityScope  string `json:"capability_scope"`
}

type Lineage struct {
	ParentEnvelopeIDs []string `json:"parent_envelope_ids"`
	GenerationDepth   int      `json:"generation_depth"`
	LinkHash          string   `json:"link_hash,omitempty"`
}

// DecodeStrict rejects oversize, duplicate-key, unknown-field, and trailing
// JSON input before assigning protocol meaning.
func DecodeStrict(reader io.Reader, maximumBytes int64) (*CyExchangeEnvelope, error) {
	if reader == nil || maximumBytes <= 0 || maximumBytes > MaxEnvelopeBytes {
		return nil, fmt.Errorf("%w: invalid decoder policy", cyerrors.ErrValidation)
	}
	raw, err := io.ReadAll(io.LimitReader(reader, maximumBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read envelope: %v", cyerrors.ErrValidation, err)
	}
	if int64(len(raw)) > maximumBytes {
		return nil, fmt.Errorf("%w: envelope exceeds byte limit", cyerrors.ErrValidation)
	}
	if err := rejectDuplicateJSONKeys(raw); err != nil {
		return nil, fmt.Errorf("%w: %v", cyerrors.ErrValidation, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var env CyExchangeEnvelope
	if err := decoder.Decode(&env); err != nil {
		return nil, fmt.Errorf("%w: decode envelope: %v", cyerrors.ErrValidation, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			err = errors.New("trailing JSON value")
		}
		return nil, fmt.Errorf("%w: %v", cyerrors.ErrValidation, err)
	}
	return &env, nil
}

// SigningBytes returns RFC 8785 bytes for every envelope field except the
// detached signature value.
func SigningBytes(env *CyExchangeEnvelope) ([]byte, error) {
	if env == nil {
		return nil, errors.New("nil envelope")
	}
	unsigned := *env
	unsigned.SignatureEd25519 = ""
	raw, err := json.Marshal(unsigned)
	if err != nil {
		return nil, err
	}
	return Canonicalize(raw)
}

// Validate checks bounded structure, payload binding, and sender-key integrity.
// It accepts 000/010 only and never admits 101 or performs Apply.
func Validate(env *CyExchangeEnvelope) error {
	if err := validateFields(env); err != nil {
		return err
	}
	expectedHash, err := HashPayload(env.PayloadBytes)
	if err != nil {
		return fmt.Errorf("%w: payload canonicalization: %v", cyerrors.ErrValidation, err)
	}
	if expectedHash != env.PayloadHash {
		return fmt.Errorf(
			"%w: payload hash mismatch: expected %s, got %s",
			cyerrors.ErrValidation,
			expectedHash,
			env.PayloadHash,
		)
	}
	signingBytes, err := SigningBytes(env)
	if err != nil {
		return fmt.Errorf("%w: signing bytes: %v", cyerrors.ErrValidation, err)
	}
	valid, err := crypto.VerifyEd25519Signature(
		env.Sender.PublicKeyEd25519,
		hex.EncodeToString(signingBytes),
		env.SignatureEd25519,
	)
	if err != nil {
		return fmt.Errorf("%w: %v", cyerrors.ErrSignature, err)
	}
	if !valid {
		return cyerrors.ErrSignature
	}
	return nil
}

func validateFields(env *CyExchangeEnvelope) error {
	if env == nil {
		return fmt.Errorf("%w: nil envelope", cyerrors.ErrValidation)
	}
	if !envelopeIDPattern.MatchString(env.EnvelopeID) {
		return fmt.Errorf("%w: invalid envelope ID", cyerrors.ErrValidation)
	}
	if env.Sender.DID == "" || env.Receiver.DID == "" ||
		env.Sender.PublicKeyEd25519 == "" || env.PayloadSchemaURI == "" {
		return fmt.Errorf("%w: required envelope field missing", cyerrors.ErrValidation)
	}
	if env.RequiredState != "000" && env.RequiredState != "010" {
		return fmt.Errorf(
			"%w: state %q is outside the 010 corridor",
			cyerrors.ErrGovernance,
			env.RequiredState,
		)
	}
	if !hashPattern.MatchString(env.PayloadHash) {
		return fmt.Errorf("%w: invalid payload hash", cyerrors.ErrValidation)
	}
	if env.SignatureEd25519 == "" {
		return fmt.Errorf("%w: signature missing", cyerrors.ErrSignature)
	}
	if err := validateLineageShape(env); err != nil {
		return err
	}
	return nil
}

func validateLineageShape(env *CyExchangeEnvelope) error {
	lineage := env.Lineage
	if lineage.GenerationDepth < 0 || len(lineage.ParentEnvelopeIDs) > 64 {
		return fmt.Errorf("%w: invalid lineage bounds", cyerrors.ErrLineage)
	}
	seen := make(map[string]struct{}, len(lineage.ParentEnvelopeIDs))
	for _, parentID := range lineage.ParentEnvelopeIDs {
		if !envelopeIDPattern.MatchString(parentID) || parentID == env.EnvelopeID {
			return fmt.Errorf("%w: invalid lineage parent", cyerrors.ErrLineage)
		}
		if _, exists := seen[parentID]; exists {
			return fmt.Errorf("%w: duplicate lineage parent", cyerrors.ErrLineage)
		}
		seen[parentID] = struct{}{}
	}
	if len(lineage.ParentEnvelopeIDs) == 0 {
		if lineage.GenerationDepth != 0 || lineage.LinkHash != "" {
			return fmt.Errorf("%w: malformed lineage root", cyerrors.ErrLineage)
		}
		return nil
	}
	if lineage.GenerationDepth == 0 || !hashPattern.MatchString(lineage.LinkHash) {
		return fmt.Errorf("%w: derived lineage binding missing", cyerrors.ErrLineage)
	}
	return nil
}

func rejectDuplicateJSONKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := readJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON token")
		}
		return err
	}
	return nil
}

func readJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = struct{}{}
			if err := readJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := readJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return errors.New("unexpected JSON delimiter")
	}
}
