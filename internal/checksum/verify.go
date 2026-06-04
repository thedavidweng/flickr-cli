package checksum

// VerifyResult represents the result of checksum verification.
type VerifyResult struct {
	Valid    int `json:"valid"`
	Missing  int `json:"missing"`
	Mismatch int `json:"mismatch"`
	Failed   int `json:"failed"`
}

// VerifyStatus represents the verification status of a single photo.
type VerifyStatus string

const (
	VerifyValid    VerifyStatus = "valid"
	VerifyMissing  VerifyStatus = "missing"
	VerifyMismatch VerifyStatus = "mismatch"
	VerifyFailed   VerifyStatus = "failed"
)

// PhotoVerifyResult is the verification result for a single photo.
type PhotoVerifyResult struct {
	PhotoID  string       `json:"photo_id"`
	Status   VerifyStatus `json:"status"`
	Expected string       `json:"expected,omitempty"`
	Actual   string       `json:"actual,omitempty"`
	Error    string       `json:"error,omitempty"`
}

// ContainsChecksum checks if a list of tags contains a checksum for the given algorithm.
// Currently only used in tests; kept as a utility for future use.
func ContainsChecksum(tags []string, algorithm string) (hex string, ok bool) {
	prefix := MachineTagPrefix + algorithm + "="
	for _, tag := range tags {
		if len(tag) > len(prefix) && tag[:len(prefix)] == prefix {
			return tag[len(prefix):], true
		}
	}
	return "", false
}
