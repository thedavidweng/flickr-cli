package upload

import "github.com/thedavidweng/flickr-cli/internal/checksum"

// PlanOptions configures the upload planner.
type PlanOptions struct {
	Recursive   bool
	AcceptedExt []string
	Dedupe      string
	Hash        string
	Tags        []string
	Albums      []string
	AlbumIDs    []string
	Description string
	Privacy     string
	Safety      string
	ContentType string
	Hidden      string
	MoveAfter   string
}

// PlannedUpload represents a planned upload action.
type PlannedUpload struct {
	LocalPath   string    `json:"local_path"`
	SizeBytes   int64     `json:"size_bytes"`
	Hash        *HashInfo `json:"hash,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Albums      []string  `json:"albums,omitempty"`
	AlbumIDs    []string  `json:"album_ids,omitempty"`
	Privacy     string    `json:"privacy,omitempty"`
	Safety      string    `json:"safety,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
	Hidden      string    `json:"hidden,omitempty"`
	Title       string    `json:"title,omitempty"`
	Desc        string    `json:"description,omitempty"`
}

// HashInfo contains computed hash information.
type HashInfo struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

// SkippedUpload represents a skipped file.
type SkippedUpload struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// Plan contains the results of upload planning.
type Plan struct {
	Planned []PlannedUpload `json:"planned"`
	Skipped []SkippedUpload `json:"skipped"`
	Invalid []InvalidFile   `json:"invalid"`
}

// BuildPlan creates an upload plan from local files.
func BuildPlan(files []LocalFile, opts *PlanOptions) (*Plan, error) {
	plan := &Plan{
		Invalid: []InvalidFile{},
		Skipped: []SkippedUpload{},
		Planned: []PlannedUpload{},
	}

	for _, f := range files {
		pu := PlannedUpload{
			LocalPath:   f.Path,
			SizeBytes:   f.Size,
			Tags:        opts.Tags,
			Albums:      opts.Albums,
			AlbumIDs:    opts.AlbumIDs,
			Privacy:     opts.Privacy,
			Safety:      opts.Safety,
			ContentType: opts.ContentType,
			Hidden:      opts.Hidden,
			Title:       f.Name,
			Desc:        opts.Description,
		}

		// Compute hash if dedupe is checksum
		if opts.Dedupe == "checksum" {
			hashVal, err := checksum.FileHash(f.Path, opts.Hash)
			if err != nil {
				plan.Skipped = append(plan.Skipped, SkippedUpload{
					Path:   f.Path,
					Reason: "hash failed: " + err.Error(),
				})
				continue
			}
			pu.Hash = &HashInfo{
				Algorithm: opts.Hash,
				Value:     hashVal,
			}
			pu.Tags = append(pu.Tags, checksum.FormatMachineTag(opts.Hash, hashVal))
		}

		plan.Planned = append(plan.Planned, pu)
	}

	return plan, nil
}
