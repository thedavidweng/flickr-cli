package flickr

import "fmt"

// PrivacyLevel represents a Flickr photo visibility level.
// It is the single source of truth for mapping user-facing strings
// ("public", "private", "friends", "family", "friends-family")
// to Flickr API parameters.
type PrivacyLevel string

const (
	PrivacyPublic       PrivacyLevel = "public"
	PrivacyPrivate      PrivacyLevel = "private"
	PrivacyFriends      PrivacyLevel = "friends"
	PrivacyFamily       PrivacyLevel = "family"
	PrivacyFriendsFamily PrivacyLevel = "friends-family"
)

// ParsePrivacyLevel converts a user-facing string to a PrivacyLevel.
// Returns an error for unrecognized values.
func ParsePrivacyLevel(s string) (PrivacyLevel, error) {
	switch s {
	case "public":
		return PrivacyPublic, nil
	case "private":
		return PrivacyPrivate, nil
	case "friends":
		return PrivacyFriends, nil
	case "family":
		return PrivacyFamily, nil
	case "friends-family":
		return PrivacyFriendsFamily, nil
	default:
		return "", fmt.Errorf("unknown privacy level %q (valid: public, private, friends, family, friends-family)", s)
	}
}

// PermsParams returns the Flickr API parameter map for flickr.photos.setPerms.
func (p PrivacyLevel) PermsParams() map[string]string {
	switch p {
	case PrivacyPublic:
		return map[string]string{"is_public": "1", "is_friend": "0", "is_family": "0"}
	case PrivacyPrivate:
		return map[string]string{"is_public": "0", "is_friend": "0", "is_family": "0"}
	case PrivacyFriends:
		return map[string]string{"is_public": "0", "is_friend": "1", "is_family": "0"}
	case PrivacyFamily:
		return map[string]string{"is_public": "0", "is_friend": "0", "is_family": "1"}
	case PrivacyFriendsFamily:
		return map[string]string{"is_public": "0", "is_friend": "1", "is_family": "1"}
	default:
		return map[string]string{"is_public": "0", "is_friend": "0", "is_family": "0"}
	}
}

// UploadFlags returns the boolean visibility flags for flickr.UploadOptions.
func (p PrivacyLevel) UploadFlags() (isPublic, isFriend, isFamily bool) {
	switch p {
	case PrivacyPublic:
		return true, false, false
	case PrivacyFriends:
		return false, true, false
	case PrivacyFamily:
		return false, false, true
	case PrivacyFriendsFamily:
		return false, true, true
	default: // PrivacyPrivate or empty
		return false, false, false
	}
}
