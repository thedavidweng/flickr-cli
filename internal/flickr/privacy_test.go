package flickr

import "testing"

func TestParsePrivacyLevel(t *testing.T) {
	tests := []struct {
		input string
		want  PrivacyLevel
		err   bool
	}{
		{"public", PrivacyPublic, false},
		{"private", PrivacyPrivate, false},
		{"friends", PrivacyFriends, false},
		{"family", PrivacyFamily, false},
		{"friends-family", PrivacyFriendsFamily, false},
		{"", "", true},
		{"foobar", "", true},
		{"PUBLIC", "", true},
	}
	for _, tc := range tests {
		got, err := ParsePrivacyLevel(tc.input)
		if tc.err {
			if err == nil {
				t.Errorf("ParsePrivacyLevel(%q): expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePrivacyLevel(%q): unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParsePrivacyLevel(%q): got %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestPrivacyLevelPermsParams(t *testing.T) {
	tests := []struct {
		level PrivacyLevel
		want  map[string]string
	}{
		{PrivacyPublic, map[string]string{"is_public": "1", "is_friend": "0", "is_family": "0"}},
		{PrivacyPrivate, map[string]string{"is_public": "0", "is_friend": "0", "is_family": "0"}},
		{PrivacyFriends, map[string]string{"is_public": "0", "is_friend": "1", "is_family": "0"}},
		{PrivacyFamily, map[string]string{"is_public": "0", "is_friend": "0", "is_family": "1"}},
		{PrivacyFriendsFamily, map[string]string{"is_public": "0", "is_friend": "1", "is_family": "1"}},
	}
	for _, tc := range tests {
		got := tc.level.PermsParams()
		for k, v := range tc.want {
			if got[k] != v {
				t.Errorf("%s.PermsParams()[%s]: got %q, want %q", tc.level, k, got[k], v)
			}
		}
	}
}

func TestPrivacyLevelUploadFlags(t *testing.T) {
	tests := []struct {
		level              PrivacyLevel
		isPublic, isFriend, isFamily bool
	}{
		{PrivacyPublic, true, false, false},
		{PrivacyPrivate, false, false, false},
		{PrivacyFriends, false, true, false},
		{PrivacyFamily, false, false, true},
		{PrivacyFriendsFamily, false, true, true},
		{"", false, false, false},
	}
	for _, tc := range tests {
		isPub, isFri, isFam := tc.level.UploadFlags()
		if isPub != tc.isPublic || isFri != tc.isFriend || isFam != tc.isFamily {
			t.Errorf("%s.UploadFlags(): got (%v,%v,%v), want (%v,%v,%v)",
				tc.level, isPub, isFri, isFam, tc.isPublic, tc.isFriend, tc.isFamily)
		}
	}
}
