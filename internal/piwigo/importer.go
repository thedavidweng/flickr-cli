package piwigo

import (
	"context"
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
)

// ImportOptions configures the Piwigo import.
type ImportOptions struct {
	Uploads     string
	DB          DBConfig
	AlbumPrefix string
	ImportAlbum string
	Dedupe      string
	Hash        string
	Limit       int
	Resume      bool
}

// ImportSummary is the result of a Piwigo import.
type ImportSummary struct {
	Planned   int `json:"planned"`
	Succeeded int `json:"succeeded"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// Importer runs the Piwigo import.
type Importer struct {
	Events    output.EventWriter
	Gate      safety.GateInput
	AuditPath string
	Profile   string
	RequestID string
}

// Import runs the Piwigo import.
func (i *Importer) Import(ctx context.Context, opts ImportOptions) (*ImportSummary, error) {
	gateResult := safety.Check(i.Gate, safety.Mutation{
		Command: "piwigo.import",
		Method:  "flickr.upload",
		Risk:    safety.RiskMediumWrite,
	})

	if gateResult.Error != nil {
		return nil, fmt.Errorf("%s", gateResult.Error.Message)
	}

	if gateResult.Planned {
		return &ImportSummary{Planned: 0}, nil
	}

	if err := opts.DB.Validate(); err != nil {
		return nil, fmt.Errorf("invalid DB config: %w", err)
	}

	return &ImportSummary{}, nil
}
