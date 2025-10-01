package songlink

import "context"

type SongLink interface {
	Link(ctx context.Context, inputURL string) (string, error)
	Info(ctx context.Context, inputURL string) (*LinkResponse, error)
}

type SonglinkImpl struct {
	base    string
	version string
}

// LinkResponse represents the top-level object returned by the Songlink API.
type LinkResponse struct {
	// The unique ID for the input entity that was supplied in the request.
	EntityUniqueID string `json:"entityUniqueId"`

	// The userCountry query param that was supplied in the request.
	UserCountry string `json:"userCountry"`

	// A URL that will render the Songlink page for this entity.
	PageURL string `json:"pageUrl"`

	// A collection of objects. Each key is a platform, and each value is an
	// object that contains data for linking to the match.
	LinksByPlatform map[string]Link `json:"linksByPlatform"`

	// A collection of objects. Each key is a unique identifier for a streaming
	// entity, and each value is an object that contains data for that entity.
	EntitiesByUniqueID map[string]Entity `json:"entitiesByUniqueId"`
}

// Link contains data for linking to a match on a specific platform.
type Link struct {
	// The unique ID for this entity.
	EntityUniqueID string `json:"entityUniqueId"`

	// The URL for this match.
	URL string `json:"url"`

	// The native app URI for mobile devices.
	NativeAppURIMobile string `json:"nativeAppUriMobile,omitempty"`

	// The native app URI for desktop devices.
	NativeAppURIDesktop string `json:"nativeAppUriDesktop,omitempty"`
}

// Entity contains data for a streaming entity, such as title, artistName, etc.
type Entity struct {
	// The unique identifier on the streaming platform/API provider.
	ID string `json:"id"`

	// The type of the entity, e.g., "song" or "album".
	Type string `json:"type"`

	// The title of the entity.
	Title string `json:"title,omitempty"`

	// The name of the artist.
	ArtistName string `json:"artistName,omitempty"`

	// The URL for the thumbnail image.
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`

	// The width of the thumbnail image.
	ThumbnailWidth int `json:"thumbnailWidth,omitempty"`

	// The height of the thumbnail image.
	ThumbnailHeight int `json:"thumbnailHeight,omitempty"`

	// The API provider that powered this match.
	APIProvider string `json:"apiProvider"`

	// An array of platforms that are "powered" by this entity.
	Platforms []string `json:"platforms"`
}

func New() SongLink {
	return &SonglinkImpl{
		base:    "https://api.song.link",
		version: "v1-alpha.1",
	}
}

func (s SonglinkImpl) baseURL() string {
	return s.base + "/" + s.version
}
