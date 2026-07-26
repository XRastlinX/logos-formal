package cyonicvalidate

import (
	"encoding/json"
	"fmt"
	"io"
)

func WriteResult(writer io.Writer, result Result, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "text":
		rejected := result.RejectedInvariant
		if rejected == "" {
			rejected = "none"
		}
		lines := [][2]string{
			{"X-Cyonic-Validator-Version", result.ValidatorVersion},
			{"X-Cubed-Bit", result.CubedBit},
			{"X-Validator-Operator", result.ValidatorOperator},
			{"X-Authority-Effect", result.AuthorityEffect},
			{"X-Validation-Profile", result.ValidationProfile},
			{"X-Metadata-Scope", result.MetadataScope},
			{"X-Target", result.Target},
			{"X-Target-Root", result.TargetRoot},
			{"X-Checks-Passed", fmt.Sprintf("%d/%d", result.Checks.Passed, result.Checks.Total)},
			{"X-Router-Decision", result.RouterDecision},
			{"X-Rejected-Invariant", rejected},
			{"X-Decision-Root", result.DecisionRoot},
		}
		for _, line := range lines {
			if _, err := fmt.Fprintf(writer, "%s: %s\n", line[0], line[1]); err != nil {
				return err
			}
		}
		for _, check := range result.Checks.Results {
			if _, err := fmt.Fprintf(writer, "%s: %s", check.ID, check.Status); err != nil {
				return err
			}
			if check.Detail != "" {
				if _, err := fmt.Fprintf(writer, " — %s", check.Detail); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(writer); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
