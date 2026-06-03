package flickr

// Endpoints holds the URLs for Flickr API access.
type Endpoints struct {
	REST         string
	Upload       string
	RequestToken string
	Authorize    string
	AccessToken  string
}

// DefaultEndpoints returns the production Flickr endpoints.
func DefaultEndpoints() Endpoints {
	return Endpoints{
		REST:         "https://www.flickr.com/services/rest/",
		Upload:       "https://up.flickr.com/services/upload/",
		RequestToken: "https://www.flickr.com/services/oauth/request_token",
		Authorize:    "https://www.flickr.com/services/oauth/authorize",
		AccessToken:  "https://www.flickr.com/services/oauth/access_token",
	}
}
