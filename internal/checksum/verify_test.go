package checksum

import "testing"

func TestVerifyResult(t *testing.T) {
	result := VerifyResult{
		Valid:    10,
		Missing:  2,
		Mismatch: 1,
		Failed:   0,
	}

	if result.Valid != 10 {
		t.Errorf("expected valid=10, got %d", result.Valid)
	}
}

func TestPhotoVerifyResult(t *testing.T) {
	result := PhotoVerifyResult{
		PhotoID:  "123",
		Status:   VerifyValid,
		Expected: "abc",
		Actual:   "abc",
	}

	if result.Status != VerifyValid {
		t.Errorf("expected valid, got %s", result.Status)
	}
}
