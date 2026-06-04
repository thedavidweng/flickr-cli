package cli

import (
	"github.com/spf13/cobra"
)

var photosCmd = &cobra.Command{
	Use:   "photos",
	Short: "Manage Flickr photos",
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
	photosDownloadCmd.Flags().String("size", "original", "size: original|large|medium|small or Flickr code: o|6k|5k|4k|3k|k|h|l|c|z|m|n|s|q|t")
	photosDownloadCmd.Flags().Int("size-max", 0, "max dimension in pixels (overrides --size)")
	photosDownloadCmd.Flags().String("metadata", "json", "metadata format: json|yaml|both|none")
	photosDownloadCmd.Flags().Bool("force", false, "overwrite existing files")
	photosDownloadCmd.Flags().String("layout", "", "directory layout: flat|album|id-dirs")
	photosDownloadCmd.Flags().StringSlice("album", nil, "album title to download (repeatable)")
	photosDownloadCmd.Flags().StringSlice("album-id", nil, "album ID to download (repeatable)")
	photosDownloadCmd.Flags().Bool("all", false, "download all albums")
	photosDownloadCmd.Flags().Bool("exif", false, "include EXIF data in metadata sidecars")

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

	photosRotateCmd.Flags().Int("degrees", 90, "rotation degrees (90, 180, 270)")

	photosListCmd.Flags().Int("page", 1, "page number")
	photosListCmd.Flags().Int("per-page", 50, "items per page")

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
}
