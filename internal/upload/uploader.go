package upload

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
)

// UploadResult is the result of uploading a single file.
type UploadResult struct {
	LocalPath string `json:"local_path"`
	PhotoID   string `json:"photo_id,omitempty"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// UploadSummary is the summary of an upload operation.
type UploadSummary struct {
	Planned   int            `json:"planned"`
	Succeeded int            `json:"succeeded"`
	Skipped   int            `json:"skipped"`
	Failed    int            `json:"failed"`
	Results   []UploadResult `json:"results"`
}

// Executor runs the upload plan.
type Executor struct {
	Client      flickr.FlickrAPI
	AuditPath   string
	Events      *output.EventWriter
	Gate        safety.GateInput
	Profile     string
	RequestID   string
	Concurrency int
	MoveAfter   string
}

// Execute runs the upload plan concurrently and returns results.
func (e *Executor) Execute(ctx context.Context, plan Plan, opts *PlanOptions) (*UploadSummary, error) {
	summary := &UploadSummary{
		Planned: len(plan.Planned),
		Results: make([]UploadResult, len(plan.Planned)),
	}

	gateResult := safety.Check(e.Gate, safety.Mutation{
		Command: "photos.upload",
		Method:  "flickr.upload",
		Risk:    safety.RiskMediumWrite,
	})

	if gateResult.Error != nil {
		return nil, fmt.Errorf("%s", gateResult.Error.Message)
	}

	if gateResult.Planned {
		for i := range plan.Planned {
			summary.Results[i] = UploadResult{
				LocalPath: plan.Planned[i].LocalPath,
				Status:    "planned",
			}
		}
		return summary, nil
	}

	if len(opts.Albums) > 0 {
		resolvedIDs, err := e.resolveAlbumNames(ctx, opts.Albums, opts.AlbumIDs)
		if err != nil {
			return nil, err
		}
		for i := range plan.Planned {
			plan.Planned[i].AlbumIDs = resolvedIDs
		}
	}

	workers := e.Concurrency
	if workers < 1 {
		workers = 1
	}

	type indexedItem struct {
		index int
		plan  *PlannedUpload
	}

	ch := make(chan indexedItem, len(plan.Planned))
	for i := range plan.Planned {
		ch <- indexedItem{index: i, plan: &plan.Planned[i]}
	}
	close(ch)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				result, err := e.uploadSingle(ctx, item.plan)
				summary.Results[item.index] = result
				if err != nil {
					mu.Lock()
					summary.Failed++
					mu.Unlock()
					continue
				}

				mu.Lock()
				switch result.Status {
				case "uploaded":
					summary.Succeeded++
				case "skipped":
					summary.Skipped++
				case "failed":
					summary.Failed++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return summary, nil
}

// resolveAlbumNames resolves album name strings to Flickr album IDs,
// merging with any explicitly provided IDs. Creates albums that don't exist.
func (e *Executor) resolveAlbumNames(ctx context.Context, names, existingIDs []string) ([]string, error) {
	resolver := NewAlbumResolver(e.Client)
	if err := resolver.Load(ctx); err != nil {
		return nil, fmt.Errorf("loading albums: %w", err)
	}

	ids := make([]string, 0, len(names)+len(existingIDs))
	ids = append(ids, existingIDs...)

	for _, name := range names {
		albumID, _, err := resolver.ResolveOrCreate(ctx, name, "")
		if err != nil {
			return nil, fmt.Errorf("resolving album %q: %w", name, err)
		}
		ids = append(ids, albumID)
	}

	return ids, nil
}

func (e *Executor) uploadSingle(ctx context.Context, pu *PlannedUpload) (UploadResult, error) {
	result := UploadResult{
		LocalPath: pu.LocalPath,
	}

	if pu.Hash != nil {
		dedup := &Deduplicator{
			Client:    e.Client,
			Algorithm: pu.Hash.Algorithm,
		}
		existingID, found, err := dedup.CheckByChecksum(ctx, pu.Hash.Value)
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("dedup check failed: %v", err)
			return result, nil
		}
		if found {
			result.Status = "skipped"
			result.PhotoID = existingID
			result.Error = "checksum already exists"
			return result, nil
		}
	}

	uploadOpts := flickr.UploadOptions{
		Title:       pu.Title,
		Description: pu.Desc,
		Tags:        pu.Tags,
	}

	if pu.Privacy != "" {
		level, err := flickr.ParsePrivacyLevel(pu.Privacy)
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("invalid privacy: %v", err)
			return result, nil
		}
		uploadOpts.IsPublic, uploadOpts.IsFriend, uploadOpts.IsFamily = level.UploadFlags()
	}

	switch pu.Safety {
	case "safe":
		uploadOpts.SafetyLevel = 1
	case "moderate":
		uploadOpts.SafetyLevel = 2
	case "restricted":
		uploadOpts.SafetyLevel = 3
	}

	switch pu.ContentType {
	case "photo":
		uploadOpts.ContentType = 1
	case "screenshot":
		uploadOpts.ContentType = 2
	case "other":
		uploadOpts.ContentType = 3
	}

	switch pu.Hidden {
	case "hidden":
		uploadOpts.Hidden = 2
	default:
		uploadOpts.Hidden = 1
	}

	uploadResult, err := e.Client.Upload(ctx, pu.LocalPath, &uploadOpts)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("upload failed: %v", err)

		if e.AuditPath != "" {
			if auditErr := safety.Append(e.AuditPath, &safety.AuditEvent{
				RequestID: e.RequestID,
				Profile:   e.Profile,
				Command:   "photos.upload",
				Method:    "flickr.upload",
				Resource:  map[string]any{"path": pu.LocalPath},
				Result:    "failed",
				Error:     err.Error(),
			}); auditErr != nil {
				return result, &model.CommandError{Code: model.ErrFilesystem, Message: fmt.Sprintf("audit write failed: %v", auditErr)}
			}
		}
		return result, nil
	}

	result.Status = "uploaded"
	result.PhotoID = uploadResult.PhotoID

	if e.MoveAfter != "" {
		if err := os.MkdirAll(e.MoveAfter, 0o755); err != nil {
			e.Events.Emit(&model.Event{
				Type:    "warning",
				Command: "photos.upload",
				Message: fmt.Sprintf("failed to create move-after dir %s: %v", e.MoveAfter, err),
			})
		} else {
			dest := filepath.Join(e.MoveAfter, filepath.Base(pu.LocalPath))
			if err := os.Rename(pu.LocalPath, dest); err != nil {
				e.Events.Emit(&model.Event{
					Type:    "warning",
					Command: "photos.upload",
					Message: fmt.Sprintf("failed to move %s to %s: %v", pu.LocalPath, dest, err),
				})
			}
		}
	}

	if e.AuditPath != "" {
		if auditErr := safety.Append(e.AuditPath, &safety.AuditEvent{
			RequestID: e.RequestID,
			Profile:   e.Profile,
			Command:   "photos.upload",
			Method:    "flickr.upload",
			Resource:  map[string]any{"path": pu.LocalPath, "photo_id": uploadResult.PhotoID},
			Result:    "success",
		}); auditErr != nil {
			return result, &model.CommandError{Code: model.ErrFilesystem, Message: fmt.Sprintf("audit write failed: %v", auditErr)}
		}
	}

	if len(pu.AlbumIDs) > 0 {
		for _, albumID := range pu.AlbumIDs {
			if err := e.Client.AddToAlbum(ctx, albumID, uploadResult.PhotoID); err != nil {
				// Log warning but don't fail the upload
				e.Events.Emit(&model.Event{
					Type:    "warning",
					Command: "photos.upload",
					Message: fmt.Sprintf("failed to add to album %s: %v", albumID, err),
				})
			}
		}
	}

	return result, nil
}
