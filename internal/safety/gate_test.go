package safety

import (
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

func TestCheckReadAllowed(t *testing.T) {
	result := Check(GateInput{}, Mutation{Risk: RiskRead})
	if !result.Allowed {
		t.Error("reads should always be allowed")
	}
	if result.Error != nil {
		t.Error("reads should have no error")
	}
}

func TestReadOnlyBlocksWrites(t *testing.T) {
	input := GateInput{ReadOnly: true}
	mutation := Mutation{Command: "photos.upload", Risk: RiskMediumWrite}

	result := Check(input, mutation)
	if result.Allowed {
		t.Error("read-only should block writes")
	}
	if result.Error == nil {
		t.Fatal("expected error")
	}
	if result.Error.Code != model.ErrReadOnlyViolation {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", result.Error.Code)
	}
}

func TestDryRunPlansWrites(t *testing.T) {
	input := GateInput{DryRun: true}
	mutation := Mutation{Command: "photos.upload", Risk: RiskMediumWrite}

	result := Check(input, mutation)
	if result.Allowed {
		t.Error("dry-run should not allow writes")
	}
	if !result.Planned {
		t.Error("dry-run should return planned=true")
	}
	if result.Error != nil {
		t.Error("dry-run should have no error")
	}
}

func TestHighRiskRequiresConfirm(t *testing.T) {
	input := GateInput{}
	mutation := Mutation{Command: "photos.delete", Risk: RiskHighWrite}

	result := Check(input, mutation)
	if result.Allowed {
		t.Error("high-risk should require confirm")
	}
	if result.Error == nil {
		t.Fatal("expected error")
	}
	if result.Error.Code != model.ErrConfirmationRequired {
		t.Errorf("expected CONFIRMATION_REQUIRED, got %s", result.Error.Code)
	}
}

func TestHighRiskWithConfirm(t *testing.T) {
	input := GateInput{Confirm: true}
	mutation := Mutation{Command: "photos.delete", Risk: RiskHighWrite}

	result := Check(input, mutation)
	if !result.Allowed {
		t.Error("high-risk with confirm should be allowed")
	}
}

func TestMediumWriteAllowed(t *testing.T) {
	input := GateInput{}
	mutation := Mutation{Command: "photos.upload", Risk: RiskMediumWrite}

	result := Check(input, mutation)
	if !result.Allowed {
		t.Error("medium writes should be allowed without confirm")
	}
}

func TestLowWriteAllowed(t *testing.T) {
	input := GateInput{}
	mutation := Mutation{Command: "albums.create", Risk: RiskMediumWrite}

	result := Check(input, mutation)
	if !result.Allowed {
		t.Error("low writes should be allowed")
	}
}

func TestClassifyRisk(t *testing.T) {
	// High-risk commands
	highRisk := []string{"photos.delete", "albums.delete", "comments.delete", "piwigo.import"}
	for _, cmd := range highRisk {
		if ClassifyRisk(cmd) != RiskHighWrite {
			t.Errorf("%s should be high write", cmd)
		}
	}

	// Medium-risk commands
	mediumRisk := []string{
		"albums.create", "albums.update",
		"photos.upload", "photos.set-meta", "photos.set-tags", "photos.add-tags",
		"photos.remove-tag", "photos.set-privacy", "photos.set-location", "photos.rotate",
		"favorites.add", "favorites.remove",
		"comments.add",
	}
	for _, cmd := range mediumRisk {
		if ClassifyRisk(cmd) != RiskMediumWrite {
			t.Errorf("%s should be medium write", cmd)
		}
	}

	// Read commands
	readCmds := []string{"photos.list", "photos.search", "photos.show", "albums.list", "albums.show", "favorites.list", "galleries.list", "groups.list", "contacts.list", "stats.popular", "api.call", "version"}
	for _, cmd := range readCmds {
		if ClassifyRisk(cmd) != RiskRead {
			t.Errorf("%s should be read", cmd)
		}
	}
}
