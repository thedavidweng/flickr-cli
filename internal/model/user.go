package model

// User represents a Flickr user.
type User struct {
	NSID      string `json:"nsid"`
	Username  string `json:"username"`
	RealName  string `json:"realname"`
	PathAlias string `json:"path_alias"`
}
