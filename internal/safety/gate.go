package safety

import "github.com/thedavidweng/flickr-cli/internal/model"

// GateInput provides the safety check parameters.
type GateInput struct {
	ReadOnly bool
	DryRun   bool
	Confirm  bool
}

// GateResult is the result of a safety gate check.
type GateResult struct {
	Allowed bool
	Planned bool
	Error   *model.ErrorBody
}

// Risk represents the risk level of a mutation.
type Risk string

const (
	RiskRead        Risk = "read"
	RiskLowWrite    Risk = "low_write"
	RiskMediumWrite Risk = "medium_write"
	RiskHighWrite   Risk = "high_write"
)

// Check evaluates a mutation against safety constraints.
func Check(input GateInput, mutation Mutation) GateResult {
	// Reads are always allowed
	if mutation.Risk == RiskRead {
		return GateResult{Allowed: true}
	}

	// Read-only blocks all writes
	if input.ReadOnly {
		return GateResult{
			Allowed: false,
			Error: &model.ErrorBody{
				Code:     model.ErrReadOnlyViolation,
				Message:  "Operation blocked by --read-only flag",
				Category: "safety",
				Details: map[string]any{
					"command": mutation.Command,
					"flag":    "--read-only",
				},
			},
		}
	}

	// Dry-run returns planned
	if input.DryRun {
		return GateResult{Planned: true, Allowed: false}
	}

	// High risk requires confirm
	if mutation.Risk == RiskHighWrite && !input.Confirm {
		return GateResult{
			Allowed: false,
			Error: &model.ErrorBody{
				Code:     model.ErrConfirmationRequired,
				Message:  "High-risk operation requires --confirm flag",
				Category: "safety",
				Details: map[string]any{
					"command": mutation.Command,
					"flag":    "--confirm",
				},
			},
		}
	}

	return GateResult{Allowed: true}
}
