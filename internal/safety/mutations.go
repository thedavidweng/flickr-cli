package safety

// HighRiskCommands lists commands that require --confirm.
var HighRiskCommands = []string{
	"photos.delete",
	"albums.delete",
	"albums.delete-photos",
	"photos.set-meta", // when batch overwriting
	"piwigo.import",   // destructive cleanup mode
}

// RemoteMutations lists commands that perform remote mutations.
var RemoteMutations = []string{
	"albums.create",
	"albums.update",
	"albums.delete",
	"photos.upload",
	"photos.delete",
	"photos.set-meta",
	"photos.set-tags",
	"photos.add-tags",
	"photos.remove-tag",
	"photos.set-privacy",
	"photos.set-location",
	"photos.rotate",
	"piwigo.import",
}

// Mutation describes a remote mutation.
type Mutation struct {
	Command  string
	Method   string
	Risk     Risk
	Resource map[string]any
}

// ClassifyRisk returns the risk level for a command.
func ClassifyRisk(command string) Risk {
	for _, c := range HighRiskCommands {
		if c == command {
			return RiskHighWrite
		}
	}
	for _, c := range RemoteMutations {
		if c == command {
			return RiskMediumWrite
		}
	}
	return RiskRead
}
