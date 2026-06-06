package flickr

// CleanContent recursively unwraps Flickr's {"_content": "value"} pattern.
// Single-key objects with only "_content" are replaced by the inner value.
// This mirrors the Python library's clean_content() helper.
func CleanContent(v any) any {
	switch val := v.(type) {
	case map[string]any:
		if len(val) == 1 {
			if content, ok := val["_content"]; ok {
				return CleanContent(content)
			}
		}
		for k, inner := range val {
			val[k] = CleanContent(inner)
		}
		return val
	case []any:
		for i, inner := range val {
			val[i] = CleanContent(inner)
		}
		return val
	default:
		return v
	}
}

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

// FlickrText is a localized text field used in many Flickr API responses.
type FlickrText struct {
	Content string `json:"_content"`
}

// PhotosetListItem is a single photoset entry from flickr.photosets.getList.
type PhotosetListItem struct {
	ID             string     `json:"id"`
	Title          FlickrText `json:"title"`
	Description    FlickrText `json:"description"`
	Photos         int        `json:"photos"`
	PrimaryPhotoID string     `json:"primary_photo_id"`
	DateCreate     string     `json:"date_create"`
	DateUpdate     string     `json:"date_update"`
}

// PhotosetListResponse is the response for flickr.photosets.getList.
type PhotosetListResponse struct {
	Photosets struct {
		Photoset []PhotosetListItem `json:"photoset"`
		Page     int                `json:"page"`
		Pages    int                `json:"pages"`
		PerPage  int                `json:"perpage"`
		Total    int                `json:"total"`
	} `json:"photosets"`
}

// DefaultExtras is the standard set of extra fields requested in pagination calls.
// This avoids per-photo API calls by fetching metadata inline.
const DefaultExtras = "description,date_upload,views,media,path_alias,owner_name,license,date_taken,geo,machine_tags,o_dims,original_format,url_o,url_k,url_l,url_m,url_s,secret,server,farm"

// PhotoListItem is a single photo entry from flickr.people.getPhotos or flickr.photos.search.
// Fields beyond ID/Title/Owner/Tags are populated when the `extras` parameter is used.
type PhotoListItem struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Owner          string  `json:"owner,omitempty"`
	Tags           string  `json:"tags,omitempty"`
	Description    string  `json:"description,omitempty"`
	DateUpload     string  `json:"dateupload,omitempty"`
	Views          string  `json:"views,omitempty"`
	Media          string  `json:"media,omitempty"`
	PathAlias      string  `json:"pathalias,omitempty"`
	OwnerName      string  `json:"ownername,omitempty"`
	License        string  `json:"license,omitempty"`
	DateTaken      string  `json:"datetaken,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	Accuracy       string  `json:"accuracy,omitempty"`
	Privacy        string  `json:"privacy,omitempty"`
	MachineTags    string  `json:"machine_tags,omitempty"`
	OriginalFormat string  `json:"originalformat,omitempty"`
	HeightO        int     `json:"height_o,omitempty"`
	WidthO         int     `json:"width_o,omitempty"`
	HeightK        int     `json:"height_k,omitempty"`
	WidthK         int     `json:"width_k,omitempty"`
	HeightL        int     `json:"height_l,omitempty"`
	WidthL         int     `json:"width_l,omitempty"`
	HeightM        int     `json:"height_m,omitempty"`
	WidthM         int     `json:"width_m,omitempty"`
	HeightS        int     `json:"height_s,omitempty"`
	WidthS         int     `json:"width_s,omitempty"`
	URLO           string  `json:"url_o,omitempty"`
	URLK           string  `json:"url_k,omitempty"`
	URLL           string  `json:"url_l,omitempty"`
	URLM           string  `json:"url_m,omitempty"`
	URLS           string  `json:"url_s,omitempty"`
	Secret         string  `json:"secret,omitempty"`
	Server         string  `json:"server,omitempty"`
	Farm           int     `json:"farm,omitempty"`
}

// PhotoListResponse is the response for flickr.people.getPhotos.
type PhotoListResponse struct {
	Photos struct {
		Photo   []PhotoListItem `json:"photo"`
		Page    int             `json:"page"`
		Pages   int             `json:"pages"`
		PerPage int             `json:"perpage"`
		Total   int             `json:"total"`
	} `json:"photos"`
}

// PhotosetGetPhotosResponse is the response for flickr.photosets.getPhotos.
type PhotosetGetPhotosResponse struct {
	Photoset struct {
		ID      string          `json:"id"`
		Photo   []PhotoListItem `json:"photo"`
		Page    int             `json:"page"`
		Pages   int             `json:"pages"`
		PerPage int             `json:"perpage"`
		Total   int             `json:"total"`
	} `json:"photoset"`
}

// ExifData holds EXIF metadata for a photo.
type ExifData struct {
	PhotoID string    `json:"id"`
	Tags    []ExifTag `json:"tag"`
}

// ExifTag is a single EXIF tag.
type ExifTag struct {
	Tagspace   string `json:"tagspace"`
	TagspaceID int    `json:"tagspaceid"`
	Tag        string `json:"tag"`
	Label      string `json:"label"`
	Raw        string `json:"raw"`
	Clean      string `json:"_content"`
}

// ExifResponse is the response for flickr.photos.getExif.
type ExifResponse struct {
	Photo ExifData `json:"photo"`
	Stat  string   `json:"stat"`
}

// PhotoSearchResponse is the response for flickr.photos.search.
type PhotoSearchResponse struct {
	Photos struct {
		Photo   []PhotoListItem `json:"photo"`
		Page    int             `json:"page"`
		Pages   int             `json:"pages"`
		PerPage int             `json:"perpage"`
		Total   int             `json:"total"`
	} `json:"photos"`
}
