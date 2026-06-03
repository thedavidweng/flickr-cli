package flickr

// RequestTokenResponse holds the response from the OAuth request_token endpoint.
type RequestTokenResponse struct {
	Token             string
	TokenSecret       string
	CallbackConfirmed bool
}

// AccessTokenResponse holds the response from the OAuth access_token endpoint.
type AccessTokenResponse struct {
	Token       string
	TokenSecret string
	UserNSID    string
	Username    string
	FullName    string
}

// LoginInfo holds the authenticated user's information returned by flickr.test.login.
type LoginInfo struct {
	UserNSID string `json:"user_nsid"`
	Username string `json:"username"`
}
