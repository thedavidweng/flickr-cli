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
