package checksum

import "strings"

const MachineTagPrefix = "checksum:"

// ParseMachineTag parses a checksum machine tag like "checksum:md5=abc123".
// Returns algorithm and hash value, or empty strings if not a valid checksum tag.
func ParseMachineTag(tag string) (algorithm, value string) {
	if !strings.HasPrefix(tag, MachineTagPrefix) {
		return "", ""
	}
	parts := strings.SplitN(strings.TrimPrefix(tag, MachineTagPrefix), "=", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", ""
	}
	return parts[0], parts[1]
}

// FormatMachineTag creates a checksum machine tag.
func FormatMachineTag(algorithm, value string) string {
	return MachineTagPrefix + algorithm + "=" + value
}
