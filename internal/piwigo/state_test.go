package piwigo

import "testing"

func TestImportStates(t *testing.T) {
	states := []ImportState{
		StatePlanned,
		StateSkippedExist,
		StateUploaded,
		StateAlbumsDone,
		StateGeoDone,
		StateDone,
		StateFailed,
	}

	for _, s := range states {
		if s == "" {
			t.Error("state should not be empty")
		}
	}
}
