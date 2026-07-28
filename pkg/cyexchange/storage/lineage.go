// Status: PROPOSED
// authority_effect: NONE
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
)

const lineageDomain = "cyexchange-lineage-link.v0.3\x00"

type ParentBinding struct {
	EnvelopeID  string
	PayloadHash string
	Depth       int
}

func ComputeLineageLink(
	childEnvelopeID, childPayloadHash string,
	parents []ParentBinding,
) string {
	copied := append([]ParentBinding(nil), parents...)
	sort.Slice(copied, func(i, j int) bool {
		return copied[i].EnvelopeID < copied[j].EnvelopeID
	})
	hash := sha256.New()
	hash.Write([]byte(lineageDomain))
	writeField(hash, childEnvelopeID)
	writeField(hash, childPayloadHash)
	for _, parent := range copied {
		writeField(hash, parent.EnvelopeID)
		writeField(hash, parent.PayloadHash)
		writeField(hash, fmt.Sprintf("%d", parent.Depth))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// VerifyLineage checks graph-local ancestry only. A valid chain records that
// the declared parents exist in this store and that the child binding matches;
// it does not establish external truth or authority.
func VerifyLineage(env *envelope.CyExchangeEnvelope, store Store) error {
	if env == nil || store == nil {
		return fmt.Errorf("%w: envelope and store are required", cyerrors.ErrLineage)
	}
	parentIDs := env.Lineage.ParentEnvelopeIDs
	if len(parentIDs) == 0 {
		if env.Lineage.GenerationDepth != 0 || env.Lineage.LinkHash != "" {
			return fmt.Errorf("%w: invalid lineage root", cyerrors.ErrLineage)
		}
		return nil
	}

	bindings := make([]ParentBinding, 0, len(parentIDs))
	seen := make(map[string]struct{}, len(parentIDs))
	maxDepth := -1
	for _, parentID := range parentIDs {
		if parentID == "" || parentID == env.EnvelopeID {
			return fmt.Errorf("%w: invalid parent identifier", cyerrors.ErrLineage)
		}
		if _, exists := seen[parentID]; exists {
			return fmt.Errorf("%w: duplicate parent identifier", cyerrors.ErrLineage)
		}
		seen[parentID] = struct{}{}
		parent, _, err := store.GetEnvelope(parentID)
		if err != nil {
			return fmt.Errorf("%w: parent %s unavailable: %v", cyerrors.ErrLineage, parentID, err)
		}
		if parent.Lineage.GenerationDepth > maxDepth {
			maxDepth = parent.Lineage.GenerationDepth
		}
		bindings = append(bindings, ParentBinding{
			EnvelopeID:  parent.EnvelopeID,
			PayloadHash: parent.PayloadHash,
			Depth:       parent.Lineage.GenerationDepth,
		})
	}
	if env.Lineage.GenerationDepth != maxDepth+1 {
		return fmt.Errorf(
			"%w: depth %d does not follow parent depth %d",
			cyerrors.ErrLineage,
			env.Lineage.GenerationDepth,
			maxDepth,
		)
	}
	expected := ComputeLineageLink(env.EnvelopeID, env.PayloadHash, bindings)
	if env.Lineage.LinkHash != expected {
		return fmt.Errorf(
			"%w: link hash mismatch: expected %s",
			cyerrors.ErrLineage,
			expected,
		)
	}
	return nil
}

type fieldWriter interface {
	Write([]byte) (int, error)
}

func writeField(writer fieldWriter, value string) {
	_, _ = writer.Write([]byte(value))
	_, _ = writer.Write([]byte{0})
}
