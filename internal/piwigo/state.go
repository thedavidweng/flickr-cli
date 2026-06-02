package piwigo

// ImportState represents the state of an import item.
type ImportState string

const (
	StatePlanned      ImportState = "planned"
	StateSkippedExist ImportState = "skipped_existing"
	StateUploaded     ImportState = "uploaded"
	StateAlbumsDone   ImportState = "albums_done"
	StateGeoDone      ImportState = "geo_done"
	StateDone         ImportState = "done"
	StateFailed       ImportState = "failed"
)
