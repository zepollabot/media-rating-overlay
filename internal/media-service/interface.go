package media

import (
	"context"
	"net/http"
	"net/url"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

// MediaResponse is a marker interface for media service responses
type MediaResponse interface {
	// Marker interface - no methods required
}

// MediaClient defines the interface for all media service clients
type MediaClient interface {
	DoWithResponse(request *http.Request) (*http.Response, error)
	DoWithMediaResponse(request *http.Request) (MediaResponse, error)
	GetBaseUrl() *url.URL
}

type FilterService interface {
	ApplyFiltersToRequest(request *http.Request, filters []model.Filter)
	ConvertConfigFiltersToRequestFilters(config *config.Library) []model.Filter
}

type PosterService interface {
	GetPosterDiskPosition(ctx context.Context, item model.Item, config *config.Library) (string, error)
	EnsurePosterExists(ctx context.Context, item model.Item, config *config.Library) error
}

type LibraryService interface {
	GetLibraries(ctx context.Context) ([]model.Library, error)
	RefreshLibrary(ctx context.Context, libraryID string, force bool) error
}

type ItemService interface {
	GetItems(ctx context.Context, library model.Library, config *config.Library) ([]model.Item, error)
}
