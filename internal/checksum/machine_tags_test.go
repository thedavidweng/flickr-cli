package checksum

import "testing"

func TestParseMachineTag(t *testing.T) {
	tests := []struct {
		tag      string
		wantAlg  string
		wantHash string
	}{
		{"checksum:md5=abc123", "md5", "abc123"},
		{"checksum:sha1=def456", "sha1", "def456"},
		{"not-a-checksum", "", ""},
		{"checksum:", "", ""},
		{"checksum:md5=", "", ""},
		{"checksum:=value", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			alg, hash := ParseMachineTag(tt.tag)
			if alg != tt.wantAlg {
				t.Errorf("algorithm = %q, want %q", alg, tt.wantAlg)
			}
			if hash != tt.wantHash {
				t.Errorf("hash = %q, want %q", hash, tt.wantHash)
			}
		})
	}
}

func TestFormatMachineTag(t *testing.T) {
	got := FormatMachineTag("md5", "abc123")
	want := "checksum:md5=abc123"
	if got != want {
		t.Errorf("FormatMachineTag = %q, want %q", got, want)
	}
}

func TestContainsChecksum(t *testing.T) {
	tags := []string{"nature", "checksum:md5=abc123", "sunset"}

	hash, ok := ContainsChecksum(tags, "md5")
	if !ok {
		t.Error("expected to find md5 checksum")
	}
	if hash != "abc123" {
		t.Errorf("expected abc123, got %s", hash)
	}

	_, ok = ContainsChecksum(tags, "sha1")
	if ok {
		t.Error("should not find sha1 checksum")
	}

	_, ok = ContainsChecksum([]string{}, "md5")
	if ok {
		t.Error("should not find checksum in empty tags")
	}
}
