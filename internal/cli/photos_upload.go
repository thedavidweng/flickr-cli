package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
	"github.com/thedavidweng/flickr-cli/internal/upload"
	"github.com/spf13/cobra"
)

var photosUploadCmd = &cobra.Command{
	Use:   "upload [path...]",
	Short: "Upload photos",
	Args:  cobra.MinimumNArgs(1),
	RunE: withAuth("photos.upload", func(ctx *CmdContext) error {
		recursive, _ := ctx.Cmd.Flags().GetBool("recursive")
		description, _ := ctx.Cmd.Flags().GetString("description")
		tagSlice, _ := ctx.Cmd.Flags().GetStringSlice("tag")
		tagsCSV, _ := ctx.Cmd.Flags().GetString("tags")
		albumNames, _ := ctx.Cmd.Flags().GetStringSlice("album")
		albumIDs, _ := ctx.Cmd.Flags().GetStringSlice("album-id")
		privacy, _ := ctx.Cmd.Flags().GetString("privacy")
		safetyLevel, _ := ctx.Cmd.Flags().GetString("safety")
		contentType, _ := ctx.Cmd.Flags().GetString("content-type")
		hidden, _ := ctx.Cmd.Flags().GetString("hidden")
		dedupe, _ := ctx.Cmd.Flags().GetString("dedupe")
		hash, _ := ctx.Cmd.Flags().GetString("hash")
		moveAfter, _ := ctx.Cmd.Flags().GetString("move-after")
		acceptedExt, _ := ctx.Cmd.Flags().GetStringSlice("accepted-ext")

		var allTags []string
		allTags = append(allTags, tagSlice...)
		if tagsCSV != "" {
			for _, t := range strings.Split(tagsCSV, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					allTags = append(allTags, t)
				}
			}
		}

		valid, invalid, err := upload.Scan(ctx.Args, upload.ScanOptions{
			Recursive:   recursive,
			AcceptedExt: acceptedExt,
		})
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFilesystem, "%v", err))
		}

		if len(valid) == 0 {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "no valid files found"))
		}

		planOpts := upload.PlanOptions{
			Recursive:   recursive,
			AcceptedExt: acceptedExt,
			Dedupe:      dedupe,
			Hash:        hash,
			Tags:        allTags,
			Albums:      albumNames,
			AlbumIDs:    albumIDs,
			Description: description,
			Privacy:     privacy,
			Safety:      safetyLevel,
			ContentType: contentType,
			Hidden:      hidden,
			MoveAfter:   moveAfter,
		}

		plan, err := upload.BuildPlan(valid, planOpts)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "%v", err))
		}

		plan.Invalid = append(plan.Invalid, invalid...)

		if ctx.App.DryRun {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"planned": true,
				"plan":    plan,
			}, nil)
		}

		if len(albumNames) > 0 {
			resolver := upload.NewAlbumResolver(ctx.Client)
			if err := resolver.Load(ctx.Cmd.Context()); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "loading albums: %v", err))
			}
			for _, name := range albumNames {
				albumID, _, err := resolver.ResolveOrCreate(ctx.Cmd.Context(), name, "")
				if err != nil {
					return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "resolving album %q: %v", name, err))
				}
				planOpts.AlbumIDs = append(planOpts.AlbumIDs, albumID)
			}
			for i := range plan.Planned {
				plan.Planned[i].AlbumIDs = planOpts.AlbumIDs
			}
		}

		auditPath := config.DefaultAuditLogPath(ctx.App.Profile)
		executor := &upload.Executor{
			Client:      ctx.Client,
			AuditPath:   auditPath,
			Events:      &output.EventWriter{Enabled: ctx.App.Events, Err: ctx.Cmd.ErrOrStderr()},
			Gate:        safety.GateInput{ReadOnly: ctx.App.ReadOnly, DryRun: ctx.App.DryRun, Confirm: ctx.App.Confirm},
			Profile:     ctx.App.Profile,
			RequestID:   ctx.App.RequestID,
			Concurrency: ctx.App.Concurrency,
			MoveAfter:   moveAfter,
		}

		summary, err := executor.Execute(ctx.Cmd.Context(), *plan, planOpts)
		if err != nil {
			var cmdErr *model.CommandError
			if errors.As(err, &cmdErr) {
				return ctx.R.Failure(ctx.Meta, output.Errorf(cmdErr.Code, "%s", cmdErr.Message))
			}
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		var warnings []string
		if len(invalid) > 0 {
			warnings = append(warnings, fmt.Sprintf("%d files skipped due to unsupported extension", len(invalid)))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, summary, warnings)
		}

		tw := output.NewTableWriter(ctx.R.Out)
		tw.Header("Path", "Status", "Photo ID", "Error")
		for _, result := range summary.Results {
			tw.Row(result.LocalPath, result.Status, result.PhotoID, result.Error)
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		ctx.R.Human("\nSummary: %d planned, %d succeeded, %d skipped, %d failed\n",
			summary.Planned, summary.Succeeded, summary.Skipped, summary.Failed)
		return nil
	}),
}
