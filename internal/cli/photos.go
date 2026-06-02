package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/thedavidweng/flickr-cli/internal/backup"
	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
	"github.com/thedavidweng/flickr-cli/internal/upload"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var photosCmd = &cobra.Command{
	Use:   "photos",
	Short: "Manage Flickr photos",
}

var photosListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your photos",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.list",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		params := map[string]string{
			"user_id":  "me",
			"page":     fmt.Sprintf("%d", page),
			"per_page": fmt.Sprintf("%d", perPage),
		}

		var result struct {
			Photos struct {
				Photo []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"photo"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"photos"`
		}

		if err := client.Call(cmd.Context(), "flickr.people.getPhotos", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		photos := make([]model.PhotoSummary, len(result.Photos.Photo))
		for i, p := range result.Photos.Photo {
			photos[i] = model.PhotoSummary{
				ID:    p.ID,
				Title: p.Title,
			}
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"items": photos,
				"pagination": map[string]any{
					"page":     result.Photos.Page,
					"pages":    result.Photos.Pages,
					"per_page": result.Photos.PerPage,
					"total":    result.Photos.Total,
				},
			}, nil)
		}

		tw := output.NewTableWriter(r.Out)
		tw.Header("ID", "Title")
		for _, p := range photos {
			tw.Row(p.ID, p.Title)
		}
		return tw.Flush()
	},
}

var photosSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search photos",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.search",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}

		params := map[string]string{}
		if userID, _ := cmd.Flags().GetString("user-id"); userID != "" {
			params["user_id"] = userID
		}
		if text, _ := cmd.Flags().GetString("text"); text != "" {
			params["text"] = text
		}
		if tags, _ := cmd.Flags().GetStringSlice("tag"); len(tags) > 0 {
			tagsStr := ""
			for i, t := range tags {
				if i > 0 {
					tagsStr += ","
				}
				tagsStr += t
			}
			params["tags"] = tagsStr
		}
		if machineTags, _ := cmd.Flags().GetStringSlice("machine-tag"); len(machineTags) > 0 {
			mtStr := ""
			for i, mt := range machineTags {
				if i > 0 {
					mtStr += ","
				}
				mtStr += mt
			}
			params["machine_tags"] = mtStr
		}
		if minUpload, _ := cmd.Flags().GetString("min-upload-date"); minUpload != "" {
			params["min_upload_date"] = minUpload
		}
		if maxUpload, _ := cmd.Flags().GetString("max-upload-date"); maxUpload != "" {
			params["max_upload_date"] = maxUpload
		}
		if minTaken, _ := cmd.Flags().GetString("min-taken-date"); minTaken != "" {
			params["min_taken_date"] = minTaken
		}
		if maxTaken, _ := cmd.Flags().GetString("max-taken-date"); maxTaken != "" {
			params["max_taken_date"] = maxTaken
		}
		if privacy, _ := cmd.Flags().GetString("privacy"); privacy != "" {
			params["privacy_filter"] = privacy
		}

		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		params["page"] = fmt.Sprintf("%d", page)
		params["per_page"] = fmt.Sprintf("%d", perPage)

		if extras, _ := cmd.Flags().GetString("extras"); extras != "" {
			params["extras"] = extras
		}

		var result struct {
			Photos struct {
				Photo []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
					Owner string `json:"owner"`
					Tags  string `json:"tags"`
				} `json:"photo"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"photos"`
		}

		if err := client.Call(cmd.Context(), "flickr.photos.search", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		photos := make([]model.PhotoSummary, len(result.Photos.Photo))
		for i, p := range result.Photos.Photo {
			photos[i] = model.PhotoSummary{
				ID:    p.ID,
				Title: p.Title,
				Owner: p.Owner,
			}
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"items": photos,
				"pagination": map[string]any{
					"page":     result.Photos.Page,
					"pages":    result.Photos.Pages,
					"per_page": result.Photos.PerPage,
					"total":    result.Photos.Total,
				},
			}, nil)
		}

		tw := output.NewTableWriter(r.Out)
		tw.Header("ID", "Title", "Owner")
		for _, p := range photos {
			tw.Row(p.ID, p.Title, p.Owner)
		}
		return tw.Flush()
	},
}

var photosShowCmd = &cobra.Command{
	Use:   "show [photo-id]",
	Short: "Show photo metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.show",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}

		photoID := args[0]
		var warnings []string

		// Get photo info
		infoParams := map[string]string{"photo_id": photoID}
		var infoResult struct {
			Photo struct {
				ID    string `json:"id"`
				Title struct {
					Content string `json:"_content"`
				} `json:"title"`
				Description struct {
					Content string `json:"_content"`
				} `json:"description"`
				Owner struct {
					NSID string `json:"nsid"`
				} `json:"owner"`
				Tags struct {
					Tag []struct {
						ID      string `json:"id"`
						Raw     string `json:"raw"`
						Machine int    `json:"machine"`
					} `json:"tag"`
				} `json:"tags"`
			} `json:"photo"`
		}

		if err := client.Call(cmd.Context(), "flickr.photos.getInfo", infoParams, &infoResult); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "getInfo: %v", err))
		}

		// Get sizes (optional)
		sizeParams := map[string]string{"photo_id": photoID}
		var sizeResult struct {
			Sizes struct {
				Size []struct {
					Label  string `json:"label"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
					Source string `json:"source"`
					URL    string `json:"url"`
				} `json:"size"`
			} `json:"sizes"`
		}
		if err := client.Call(cmd.Context(), "flickr.photos.getSizes", sizeParams, &sizeResult); err != nil {
			warnings = append(warnings, fmt.Sprintf("getSizes failed: %v", err))
		}

		// Get contexts (optional)
		ctxParams := map[string]string{"photo_id": photoID}
		var ctxResult struct {
			Set []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"set"`
		}
		if err := client.Call(cmd.Context(), "flickr.photos.getAllContexts", ctxParams, &ctxResult); err != nil {
			warnings = append(warnings, fmt.Sprintf("getAllContexts failed: %v", err))
		}

		photo := map[string]any{
			"id":          infoResult.Photo.ID,
			"title":       infoResult.Photo.Title.Content,
			"description": infoResult.Photo.Description.Content,
			"owner":       infoResult.Photo.Owner.NSID,
		}

		tags := make([]string, len(infoResult.Photo.Tags.Tag))
		for i, t := range infoResult.Photo.Tags.Tag {
			tags[i] = t.Raw
		}
		photo["tags"] = tags

		if len(sizeResult.Sizes.Size) > 0 {
			sizes := make([]map[string]any, len(sizeResult.Sizes.Size))
			for i, s := range sizeResult.Sizes.Size {
				sizes[i] = map[string]any{
					"label":  s.Label,
					"width":  s.Width,
					"height": s.Height,
					"source": s.Source,
					"url":    s.URL,
				}
			}
			photo["sizes"] = sizes
		}

		if len(ctxResult.Set) > 0 {
			albums := make([]map[string]string, len(ctxResult.Set))
			for i, s := range ctxResult.Set {
				albums[i] = map[string]string{"id": s.ID, "title": s.Title}
			}
			photo["albums"] = albums
		}

		return r.Success(meta, photo, warnings)
	},
}

var photosUploadCmd = &cobra.Command{
	Use:   "upload [path...]",
	Short: "Upload photos",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.upload",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		// 1. Get client, require auth
		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		// 2. Read flags
		recursive, _ := cmd.Flags().GetBool("recursive")
		description, _ := cmd.Flags().GetString("description")
		tagSlice, _ := cmd.Flags().GetStringSlice("tag")
		tagsCSV, _ := cmd.Flags().GetString("tags")
		albumNames, _ := cmd.Flags().GetStringSlice("album")
		albumIDs, _ := cmd.Flags().GetStringSlice("album-id")
		privacy, _ := cmd.Flags().GetString("privacy")
		safetyLevel, _ := cmd.Flags().GetString("safety")
		contentType, _ := cmd.Flags().GetString("content-type")
		hidden, _ := cmd.Flags().GetString("hidden")
		dedupe, _ := cmd.Flags().GetString("dedupe")
		hash, _ := cmd.Flags().GetString("hash")
		moveAfter, _ := cmd.Flags().GetString("move-after")
		acceptedExt, _ := cmd.Flags().GetStringSlice("accepted-ext")

		// Merge --tag and --tags into a single list
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

		// 3. Scan files
		valid, invalid, err := upload.Scan(args, upload.ScanOptions{
			Recursive:   recursive,
			AcceptedExt: acceptedExt,
		})
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFilesystem, "%v", err))
		}

		if len(valid) == 0 {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "no valid files found"))
		}

		// 4. Build plan
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
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "%v", err))
		}

		// Append invalid files to the plan
		plan.Invalid = append(plan.Invalid, invalid...)

		// 5. Dry-run: return plan as JSON
		if app.DryRun {
			return r.Success(meta, map[string]any{
				"planned": true,
				"plan":    plan,
			}, nil)
		}

		// 6. Create AlbumResolver if albums are specified
		if len(albumNames) > 0 {
			resolver := upload.NewAlbumResolver(client)
			if err := resolver.Load(cmd.Context()); err != nil {
				return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "loading albums: %v", err))
			}
			for _, name := range albumNames {
				albumID, _, err := resolver.ResolveOrCreate(cmd.Context(), name, "")
				if err != nil {
					return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "resolving album %q: %v", name, err))
				}
				planOpts.AlbumIDs = append(planOpts.AlbumIDs, albumID)
			}
			// Update planned uploads with resolved album IDs
			for i := range plan.Planned {
				plan.Planned[i].AlbumIDs = planOpts.AlbumIDs
			}
		}

		// 7. Create executor
		auditPath := config.DefaultAuditLogPath(app.Profile)
		executor := &upload.Executor{
			Client:      client,
			AuditPath:   auditPath,
			Events:      output.EventWriter{Enabled: app.Events, Err: cmd.ErrOrStderr()},
			Gate:        safety.GateInput{ReadOnly: app.ReadOnly, DryRun: app.DryRun, Confirm: app.Confirm},
			Profile:     app.Profile,
			RequestID:   app.RequestID,
			Concurrency: app.Concurrency,
		}

		// 8. Execute
		summary, err := executor.Execute(cmd.Context(), *plan, planOpts)
		if err != nil {
			var cmdErr *model.CommandError
			if errors.As(err, &cmdErr) {
				return r.Failure(meta, output.Errorf(cmdErr.Code, "%s", cmdErr.Message))
			}
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		// 9. Return summary
		var warnings []string
		if len(invalid) > 0 {
			warnings = append(warnings, fmt.Sprintf("%d files skipped due to unsupported extension", len(invalid)))
		}

		if app.JSON {
			return r.Success(meta, summary, warnings)
		}

		// Human-readable table
		tw := output.NewTableWriter(r.Out)
		tw.Header("Path", "Status", "Photo ID", "Error")
		for _, result := range summary.Results {
			tw.Row(result.LocalPath, result.Status, result.PhotoID, result.Error)
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		r.Human("\nSummary: %d planned, %d succeeded, %d skipped, %d failed\n",
			summary.Planned, summary.Succeeded, summary.Skipped, summary.Failed)
		return nil
	},
}

var photosDownloadCmd = &cobra.Command{
	Use:   "download [photo-id...]",
	Short: "Download photos",
	Long: `Download photos from Flickr. Supports three modes:

1. By photo ID: flickr photos download 12345 67890
2. By album: flickr photos download --album "Vacation" --dest ./backup
3. All photos: flickr photos download --all --dest ./backup --layout id-dirs`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.download",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		dest, _ := cmd.Flags().GetString("dest")
		size, _ := cmd.Flags().GetString("size")
		metadata, _ := cmd.Flags().GetString("metadata")
		force, _ := cmd.Flags().GetBool("force")
		layout, _ := cmd.Flags().GetString("layout")
		albumTitles, _ := cmd.Flags().GetStringSlice("album")
		albumIDs, _ := cmd.Flags().GetStringSlice("album-id")
		allAlbums, _ := cmd.Flags().GetBool("all")
		resume, _ := cmd.Flags().GetBool("resume")

		if dest == "" {
			dest = "./flickr-backup"
		}

		// Determine mode: backup mode (album/all) vs direct download mode (photo IDs)
		backupMode := allAlbums || len(albumTitles) > 0 || len(albumIDs) > 0 || layout != ""

		if backupMode {
			// Backup mode: use backup package
			return downloadViaBackup(cmd, client, r, meta, app, backup.BackupPlanOptions{
				Mode:          backupModeToPlanMode(layout, allAlbums, len(albumTitles) > 0 || len(albumIDs) > 0),
				Dest:          dest,
				AlbumTitles:   albumTitles,
				AlbumIDs:      albumIDs,
				All:           allAlbums,
				Size:          size,
				Metadata:      metadata,
				Force:         force,
				Resume:        resume,
			})
		}

		// Direct download mode: download specific photo IDs
		if len(args) == 0 {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "specify photo IDs, --album, --album-id, or --all"))
		}

		return downloadByIDs(cmd, client, r, meta, app, args, dest, size, metadata, force)
	},
}

// backupModeToPlanMode converts CLI flags to backup plan mode.
func backupModeToPlanMode(layout string, all bool, hasAlbums bool) backup.PlanMode {
	switch layout {
	case "id-dirs":
		return backup.BackupIDDirs
	case "album":
		return backup.BackupAlbums
	default:
		if all || hasAlbums {
			return backup.BackupAlbums
		}
		return backup.BackupUser
	}
}

// downloadViaBackup handles backup mode using the backup package.
func downloadViaBackup(cmd *cobra.Command, client *flickr.Client, r output.Renderer, meta output.RuntimeMetaInput, app *AppContext, opts backup.BackupPlanOptions) error {
	plan, err := backup.BuildPlan(cmd.Context(), client, opts)
	if err != nil {
		return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "%v", err))
	}

	if len(plan.Items) == 0 {
		r.Human("No photos to download\n")
		return r.Success(meta, map[string]any{"total": 0}, nil)
	}

	if app.DryRun {
		r.Human("Would download %d photos to %s\n", len(plan.Items), opts.Dest)
		return r.Success(meta, map[string]any{"planned": true, "total": len(plan.Items)}, nil)
	}

	// Build download items from plan
	items := make([]backup.DownloadItem, len(plan.Items))
	for i, item := range plan.Items {
		ext := "jpg" // Default extension
		var filePath string

		switch opts.Mode {
		case backup.BackupIDDirs:
			filePath = backup.IDDirsPath(opts.Dest, item.PhotoID, ext)
		default:
			fileName := backup.SafeName(item.Title, item.PhotoID) + "." + ext
			if item.AlbumID != "" {
				albumName := backup.SafeName(item.Title, item.AlbumID)
				filePath = filepath.Join(opts.Dest, albumName, fileName)
			} else {
				filePath = filepath.Join(opts.Dest, fileName)
			}
		}

		items[i] = backup.DownloadItem{
			PhotoID:  item.PhotoID,
			FilePath: filePath,
			SizeLabel: opts.Size,
		}
	}

	// Download
	downloader := &backup.Downloader{
		HTTP:        &http.Client{},
		Client:      client,
		Concurrency: app.Concurrency,
		Events:      output.EventWriter{Enabled: app.Events, Err: cmd.ErrOrStderr()},
	}

	summary, err := downloader.Download(cmd.Context(), items, backup.DownloadOptions{
		Force:    opts.Force,
		Size:     opts.Size,
		Metadata: opts.Metadata,
	})
	if err != nil {
		return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
	}

	if app.JSON {
		return r.Success(meta, map[string]any{
			"summary": summary,
			"dest":    opts.Dest,
		}, nil)
	}

	r.Human("\nSummary: %d total, %d completed, %d skipped, %d failed\n",
		summary.Total, summary.Completed, summary.Skipped, summary.Failed)
	return nil
}

// downloadByIDs handles direct download of specific photo IDs.
func downloadByIDs(cmd *cobra.Command, client *flickr.Client, r output.Renderer, meta output.RuntimeMetaInput, app *AppContext, photoIDs []string, dest, size, metadata string, force bool) error {
	if app.DryRun {
		var planned []map[string]any
		for _, id := range photoIDs {
			planned = append(planned, map[string]any{"photo_id": id, "dest": dest, "size": size})
		}
		r.Human("Would download %d photos to %s\n", len(photoIDs), dest)
		return r.Success(meta, map[string]any{"planned": true, "items": planned}, nil)
	}

	type downloadResult struct {
		PhotoID string `json:"photo_id"`
		Path    string `json:"path,omitempty"`
		Status  string `json:"status"`
		Error   string `json:"error,omitempty"`
	}

	results := make([]downloadResult, len(photoIDs))
	summary := map[string]int{"total": len(photoIDs), "completed": 0, "skipped": 0, "failed": 0}

	httpClient := &http.Client{}

	for i, photoID := range photoIDs {
		results[i] = downloadResult{PhotoID: photoID}

		// Get sizes
		sizes, err := client.GetSizes(cmd.Context(), photoID)
		if err != nil {
			results[i].Status = "failed"
			results[i].Error = fmt.Sprintf("getting sizes: %v", err)
			summary["failed"]++
			continue
		}

		selected, err := flickr.SelectSize(sizes, size)
		if err != nil {
			results[i].Status = "failed"
			results[i].Error = fmt.Sprintf("selecting size: %v", err)
			summary["failed"]++
			continue
		}

		// Determine file extension from URL
		ext := filepath.Ext(selected.Source)
		if ext == "" {
			ext = ".jpg"
		}
		filePath := filepath.Join(dest, photoID+ext)

		// Check if file exists
		if !force {
			if _, err := os.Stat(filePath); err == nil {
				results[i].Status = "skipped"
				results[i].Path = filePath
				results[i].Error = "file exists"
				summary["skipped"]++
				continue
			}
		}

		// Create directory
		if err := os.MkdirAll(dest, 0o755); err != nil {
			results[i].Status = "failed"
			results[i].Error = fmt.Sprintf("creating dir: %v", err)
			summary["failed"]++
			continue
		}

		// Download file
		resp, err := httpClient.Get(selected.Source)
		if err != nil {
			results[i].Status = "failed"
			results[i].Error = fmt.Sprintf("downloading: %v", err)
			summary["failed"]++
			continue
		}

		// Write to temp file then rename (atomic)
		tmpPath := filePath + ".tmp"
		f, err := os.Create(tmpPath)
		if err != nil {
			resp.Body.Close()
			results[i].Status = "failed"
			results[i].Error = fmt.Sprintf("creating file: %v", err)
			summary["failed"]++
			continue
		}

		if _, err := io.Copy(f, resp.Body); err != nil {
			f.Close()
			resp.Body.Close()
			os.Remove(tmpPath)
			results[i].Status = "failed"
			results[i].Error = fmt.Sprintf("writing file: %v", err)
			summary["failed"]++
			continue
		}
		f.Close()
		resp.Body.Close()

		if err := os.Rename(tmpPath, filePath); err != nil {
			os.Remove(tmpPath)
			results[i].Status = "failed"
			results[i].Error = fmt.Sprintf("renaming: %v", err)
			summary["failed"]++
			continue
		}

		results[i].Status = "completed"
		results[i].Path = filePath
		summary["completed"]++

		// Write metadata sidecar if requested
		if metadata == "json" || metadata == "yaml" || metadata == "both" {
			infoParams := map[string]string{"photo_id": photoID}
			var infoResult map[string]any
			if err := client.Call(cmd.Context(), "flickr.photos.getInfo", infoParams, &infoResult); err == nil {
				if metadata == "json" || metadata == "both" {
					metaPath := filePath + ".json"
					if metaBytes, mErr := json.MarshalIndent(infoResult, "", "  "); mErr == nil {
						os.WriteFile(metaPath, metaBytes, 0o644)
					}
				}
				if metadata == "yaml" || metadata == "both" {
					metaPath := filePath + ".yaml"
					if metaBytes, mErr := yaml.Marshal(infoResult); mErr == nil {
						os.WriteFile(metaPath, metaBytes, 0o644)
					}
				}
			}
		}
	}

	if app.JSON {
		return r.Success(meta, map[string]any{
			"results": results,
			"summary": summary,
			"dest":    dest,
			"size":    size,
		}, nil)
	}

	tw := output.NewTableWriter(r.Out)
	tw.Header("Photo ID", "Status", "Path", "Error")
	for _, res := range results {
		tw.Row(res.PhotoID, res.Status, res.Path, res.Error)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	r.Human("\nSummary: %d total, %d completed, %d skipped, %d failed\n",
		summary["total"], summary["completed"], summary["skipped"], summary["failed"])
	return nil
}

var photosDeleteCmd = &cobra.Command{
	Use:   "delete [photo-id...]",
	Short: "Delete photos",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.delete",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		gate := safety.Check(safety.GateInput{
			ReadOnly: app.ReadOnly,
			DryRun:   app.DryRun,
			Confirm:  app.Confirm,
		}, safety.Mutation{
			Command: "photos.delete",
			Method:  "flickr.photos.delete",
			Risk:    safety.ClassifyRisk("photos.delete"),
		})
		if gate.Error != nil {
			return r.Failure(meta, *gate.Error)
		}
		if gate.Planned {
			r.Human("Would delete %d photo(s): %s\n", len(args), strings.Join(args, ", "))
			return r.Success(meta, map[string]any{"planned": true, "photo_ids": args}, nil)
		}

		var deleted []string
		for _, id := range args {
			params := map[string]string{"photo_id": id}
			if err := client.Call(cmd.Context(), "flickr.photos.delete", params, nil); err != nil {
				return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "delete %s: %v", id, err))
			}
			deleted = append(deleted, id)
		}

		r.Human("Deleted %d photo(s): %s\n", len(deleted), strings.Join(deleted, ", "))
		return r.Success(meta, map[string]any{"deleted": deleted}, nil)
	},
}

var photosSetMetaCmd = &cobra.Command{
	Use:   "set-meta [photo-id]",
	Short: "Set photo title and description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.set-meta",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		gate := safety.Check(safety.GateInput{
			ReadOnly: app.ReadOnly,
			DryRun:   app.DryRun,
			Confirm:  app.Confirm,
		}, safety.Mutation{
			Command: "photos.set-meta",
			Method:  "flickr.photos.setMeta",
			Risk:    safety.ClassifyRisk("photos.set-meta"),
		})
		if gate.Error != nil {
			return r.Failure(meta, *gate.Error)
		}
		if gate.Planned {
			photoID := args[0]
			title, _ := cmd.Flags().GetString("title")
			description, _ := cmd.Flags().GetString("description")
			r.Human("Would set metadata on photo %s (title=%q, description=%q)\n", photoID, title, description)
			return r.Success(meta, map[string]any{"planned": true, "photo_id": photoID, "title": title, "description": description}, nil)
		}

		photoID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		params := map[string]string{
			"photo_id": photoID,
			"title":    title,
		}
		if description != "" {
			params["description"] = description
		}

		if err := client.Call(cmd.Context(), "flickr.photos.setMeta", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Updated metadata for photo %s\n", photoID)
		return r.Success(meta, map[string]any{"photo_id": photoID}, nil)
	},
}

var photosSetTagsCmd = &cobra.Command{
	Use:   "set-tags [photo-id]",
	Short: "Set photo tags (replaces existing)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.set-tags",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		gate := safety.Check(safety.GateInput{
			ReadOnly: app.ReadOnly,
			DryRun:   app.DryRun,
			Confirm:  app.Confirm,
		}, safety.Mutation{
			Command: "photos.set-tags",
			Method:  "flickr.photos.setTags",
			Risk:    safety.ClassifyRisk("photos.set-tags"),
		})
		if gate.Error != nil {
			return r.Failure(meta, *gate.Error)
		}
		if gate.Planned {
			r.Human("Would set tags on photo %s\n", args[0])
			return r.Success(meta, map[string]any{"planned": true, "photo_id": args[0]}, nil)
		}

		photoID := args[0]
		tagSlice, _ := cmd.Flags().GetStringSlice("tag")
		csvTags, _ := cmd.Flags().GetString("tags")
		if csvTags != "" {
			for _, t := range strings.Split(csvTags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tagSlice = append(tagSlice, t)
				}
			}
		}
		tags := strings.Join(tagSlice, " ")

		params := map[string]string{
			"photo_id": photoID,
			"tags":     tags,
		}

		if err := client.Call(cmd.Context(), "flickr.photos.setTags", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Set tags on photo %s\n", photoID)
		return r.Success(meta, map[string]any{"photo_id": photoID, "tags": tagSlice}, nil)
	},
}

var photosAddTagsCmd = &cobra.Command{
	Use:   "add-tags [photo-id]",
	Short: "Add tags to photo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.add-tags",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		gate := safety.Check(safety.GateInput{
			ReadOnly: app.ReadOnly,
			DryRun:   app.DryRun,
			Confirm:  app.Confirm,
		}, safety.Mutation{
			Command: "photos.add-tags",
			Method:  "flickr.photos.addTags",
			Risk:    safety.ClassifyRisk("photos.add-tags"),
		})
		if gate.Error != nil {
			return r.Failure(meta, *gate.Error)
		}
		if gate.Planned {
			r.Human("Would add tags to photo %s\n", args[0])
			return r.Success(meta, map[string]any{"planned": true, "photo_id": args[0]}, nil)
		}

		photoID := args[0]
		tagSlice, _ := cmd.Flags().GetStringSlice("tag")
		csvTags, _ := cmd.Flags().GetString("tags")
		if csvTags != "" {
			for _, t := range strings.Split(csvTags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tagSlice = append(tagSlice, t)
				}
			}
		}
		tags := strings.Join(tagSlice, " ")

		params := map[string]string{
			"photo_id": photoID,
			"tags":     tags,
		}

		if err := client.Call(cmd.Context(), "flickr.photos.addTags", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Added tags to photo %s\n", photoID)
		return r.Success(meta, map[string]any{"photo_id": photoID, "tags": tagSlice}, nil)
	},
}

var photosRemoveTagCmd = &cobra.Command{
	Use:   "remove-tag [photo-id]",
	Short: "Remove a tag from photo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.remove-tag",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		gate := safety.Check(safety.GateInput{
			ReadOnly: app.ReadOnly,
			DryRun:   app.DryRun,
			Confirm:  app.Confirm,
		}, safety.Mutation{
			Command: "photos.remove-tag",
			Method:  "flickr.photos.removeTag",
			Risk:    safety.ClassifyRisk("photos.remove-tag"),
		})
		if gate.Error != nil {
			return r.Failure(meta, *gate.Error)
		}
		if gate.Planned {
			tagID, _ := cmd.Flags().GetString("tag-id")
			r.Human("Would remove tag %s from photo %s\n", tagID, args[0])
			return r.Success(meta, map[string]any{"planned": true, "photo_id": args[0], "tag_id": tagID}, nil)
		}

		photoID := args[0]
		tagID, _ := cmd.Flags().GetString("tag-id")

		params := map[string]string{
			"photo_id": photoID,
			"tag_id":   tagID,
		}

		if err := client.Call(cmd.Context(), "flickr.photos.removeTag", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Removed tag %s from photo %s\n", tagID, photoID)
		return r.Success(meta, map[string]any{"photo_id": photoID, "tag_id": tagID}, nil)
	},
}

var photosSetPrivacyCmd = &cobra.Command{
	Use:   "set-privacy [photo-id]",
	Short: "Set photo privacy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.set-privacy",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		photoID := args[0]

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		gateResult := safety.Check(safety.GateInput{ReadOnly: app.ReadOnly, DryRun: app.DryRun, Confirm: app.Confirm}, safety.Mutation{
			Command: "photos.set-privacy",
			Method:  "flickr.photos.setPerms",
			Risk:    safety.ClassifyRisk("photos.set-privacy"),
		})
		if gateResult.Error != nil {
			return r.Failure(meta, *gateResult.Error)
		}
		if gateResult.Planned {
			r.Human("Would set privacy for photo %s\n", photoID)
			return r.Success(meta, map[string]any{"planned": true, "photo_id": photoID}, nil)
		}

		privacy, _ := cmd.Flags().GetString("privacy")
		hidden, _ := cmd.Flags().GetString("hidden")

		params := map[string]string{"photo_id": photoID}
		switch privacy {
		case "public":
			params["is_public"] = "1"
			params["is_friend"] = "0"
			params["is_family"] = "0"
		case "private":
			params["is_public"] = "0"
			params["is_friend"] = "0"
			params["is_family"] = "0"
		case "friends":
			params["is_public"] = "0"
			params["is_friend"] = "1"
			params["is_family"] = "0"
		case "family":
			params["is_public"] = "0"
			params["is_friend"] = "0"
			params["is_family"] = "1"
		case "friends-family":
			params["is_public"] = "0"
			params["is_friend"] = "1"
			params["is_family"] = "1"
		}
		if hidden == "hidden" {
			params["hidden"] = "2"
		} else if hidden == "public" {
			params["hidden"] = "1"
		}

		if err := client.Call(cmd.Context(), "flickr.photos.setPerms", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Set privacy for photo %s\n", photoID)
		return r.Success(meta, map[string]any{"photo_id": photoID, "privacy": privacy}, nil)
	},
}

var photosSetLocationCmd = &cobra.Command{
	Use:   "set-location [photo-id]",
	Short: "Set photo location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.set-location",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		photoID := args[0]

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		gateResult := safety.Check(safety.GateInput{ReadOnly: app.ReadOnly, DryRun: app.DryRun, Confirm: app.Confirm}, safety.Mutation{
			Command: "photos.set-location",
			Method:  "flickr.photos.geo.setLocation",
			Risk:    safety.ClassifyRisk("photos.set-location"),
		})
		if gateResult.Error != nil {
			return r.Failure(meta, *gateResult.Error)
		}
		if gateResult.Planned {
			r.Human("Would set location for photo %s\n", photoID)
			return r.Success(meta, map[string]any{"planned": true, "photo_id": photoID}, nil)
		}

		lat, _ := cmd.Flags().GetFloat64("lat")
		lon, _ := cmd.Flags().GetFloat64("lon")
		accuracy, _ := cmd.Flags().GetInt("accuracy")
		context, _ := cmd.Flags().GetInt("context")

		params := map[string]string{
			"photo_id": photoID,
			"lat":      fmt.Sprintf("%f", lat),
			"lon":      fmt.Sprintf("%f", lon),
		}
		if accuracy > 0 {
			params["accuracy"] = fmt.Sprintf("%d", accuracy)
		}
		if context > 0 {
			params["context"] = fmt.Sprintf("%d", context)
		}

		if err := client.Call(cmd.Context(), "flickr.photos.geo.setLocation", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Set location for photo %s\n", photoID)
		return r.Success(meta, map[string]any{"photo_id": photoID, "lat": lat, "lon": lon}, nil)
	},
}

var photosRotateCmd = &cobra.Command{
	Use:   "rotate [photo-id]",
	Short: "Rotate photo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "photos.rotate",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		photoID := args[0]

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		gateResult := safety.Check(safety.GateInput{ReadOnly: app.ReadOnly, DryRun: app.DryRun, Confirm: app.Confirm}, safety.Mutation{
			Command: "photos.rotate",
			Method:  "flickr.photos.transform.rotate",
			Risk:    safety.ClassifyRisk("photos.rotate"),
		})
		if gateResult.Error != nil {
			return r.Failure(meta, *gateResult.Error)
		}
		if gateResult.Planned {
			r.Human("Would rotate photo %s\n", photoID)
			return r.Success(meta, map[string]any{"planned": true, "photo_id": photoID}, nil)
		}

		degrees, _ := cmd.Flags().GetInt("degrees")
		if degrees == 0 {
			degrees = 90
		}

		params := map[string]string{
			"photo_id": photoID,
			"degrees":  fmt.Sprintf("%d", degrees),
		}

		if err := client.Call(cmd.Context(), "flickr.photos.transform.rotate", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Rotated photo %s by %d°\n", photoID, degrees)
		return r.Success(meta, map[string]any{"photo_id": photoID, "degrees": degrees}, nil)
	},
}

func init() {
	photosSearchCmd.Flags().String("user-id", "", "user ID or 'me'")
	photosSearchCmd.Flags().String("text", "", "search text")
	photosSearchCmd.Flags().StringSlice("tag", nil, "filter by tag (repeatable)")
	photosSearchCmd.Flags().StringSlice("machine-tag", nil, "filter by machine tag (repeatable)")
	photosSearchCmd.Flags().String("min-upload-date", "", "minimum upload date")
	photosSearchCmd.Flags().String("max-upload-date", "", "maximum upload date")
	photosSearchCmd.Flags().String("min-taken-date", "", "minimum taken date")
	photosSearchCmd.Flags().String("max-taken-date", "", "maximum taken date")
	photosSearchCmd.Flags().String("privacy", "", "privacy level filter")
	photosSearchCmd.Flags().String("album-id", "", "filter by album ID")
	photosSearchCmd.Flags().Int("page", 1, "page number")
	photosSearchCmd.Flags().Int("per-page", 50, "items per page")
	photosSearchCmd.Flags().String("extras", "", "extra fields CSV")

	photosUploadCmd.Flags().Bool("recursive", false, "recurse into directories")
	photosUploadCmd.Flags().String("description", "", "description for uploaded files")
	photosUploadCmd.Flags().StringSlice("tag", nil, "user tag (repeatable)")
	photosUploadCmd.Flags().String("tags", "", "CSV tag input")
	photosUploadCmd.Flags().StringSlice("album", nil, "album name (repeatable, created when absent)")
	photosUploadCmd.Flags().StringSlice("album-id", nil, "existing album ID (repeatable)")
	photosUploadCmd.Flags().String("privacy", "", "privacy: public|private|friends|family|friends-family")
	photosUploadCmd.Flags().String("safety", "", "safety level: safe|moderate|restricted")
	photosUploadCmd.Flags().String("content-type", "", "content type: photo|screenshot|other")
	photosUploadCmd.Flags().String("hidden", "", "hidden: public|hidden")
	photosUploadCmd.Flags().String("dedupe", "checksum", "deduplication: none|checksum")
	photosUploadCmd.Flags().String("hash", "md5", "hash algorithm: md5|sha1")
	photosUploadCmd.Flags().String("move-after", "", "move files after upload")
	photosUploadCmd.Flags().StringSlice("accepted-ext", nil, "accepted file extensions")

	photosDownloadCmd.Flags().String("dest", "", "destination directory (default ./flickr-backup)")
	photosDownloadCmd.Flags().String("size", "original", "size: original|large|medium|small")
	photosDownloadCmd.Flags().String("metadata", "json", "metadata format: json|yaml|both|none")
	photosDownloadCmd.Flags().Bool("force", false, "overwrite existing files")
	photosDownloadCmd.Flags().String("layout", "", "directory layout: flat|album|id-dirs")
	photosDownloadCmd.Flags().StringSlice("album", nil, "album title to download (repeatable)")
	photosDownloadCmd.Flags().StringSlice("album-id", nil, "album ID to download (repeatable)")
	photosDownloadCmd.Flags().Bool("all", false, "download all albums")
	photosDownloadCmd.Flags().Bool("resume", false, "resume interrupted download")

	photosSetMetaCmd.Flags().String("title", "", "photo title")
	photosSetMetaCmd.Flags().String("description", "", "photo description")

	photosSetTagsCmd.Flags().StringSlice("tag", nil, "tag (repeatable)")
	photosSetTagsCmd.Flags().String("tags", "", "CSV tags")

	photosAddTagsCmd.Flags().StringSlice("tag", nil, "tag (repeatable)")
	photosAddTagsCmd.Flags().String("tags", "", "CSV tags")

	photosRemoveTagCmd.Flags().String("tag-id", "", "tag ID to remove")

	photosSetPrivacyCmd.Flags().String("privacy", "", "privacy: public|private|friends|family|friends-family")
	photosSetPrivacyCmd.Flags().String("hidden", "", "hidden: public|hidden")

	photosSetLocationCmd.Flags().Float64("lat", 0, "latitude")
	photosSetLocationCmd.Flags().Float64("lon", 0, "longitude")
	photosSetLocationCmd.Flags().Int("accuracy", 0, "accuracy (1-16)")
	photosSetLocationCmd.Flags().Int("context", 0, "context (0=not set, 1=indoor, 2=outdoor)")

	photosCmd.AddCommand(photosListCmd)
	photosCmd.AddCommand(photosSearchCmd)
	photosCmd.AddCommand(photosShowCmd)
	photosCmd.AddCommand(photosUploadCmd)
	photosCmd.AddCommand(photosDownloadCmd)
	photosCmd.AddCommand(photosDeleteCmd)
	photosCmd.AddCommand(photosSetMetaCmd)
	photosCmd.AddCommand(photosSetTagsCmd)
	photosCmd.AddCommand(photosAddTagsCmd)
	photosCmd.AddCommand(photosRemoveTagCmd)
	photosCmd.AddCommand(photosSetPrivacyCmd)
	photosCmd.AddCommand(photosSetLocationCmd)
	photosCmd.AddCommand(photosRotateCmd)

	_ = fmt.Sprintf("photos commands registered")
}
