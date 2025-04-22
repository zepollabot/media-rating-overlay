package rating

import (
	"context"
	"net/http"
	"net/url"

	"github.com/zepollabot/media-rating-overlay/internal/model"
)

// RatingResponse is a marker interface for rating service responses
type RatingResponse interface {
}

// RatingClient defines the interface for all rating service clients
type RatingClient interface {
	DoWithResponse(request *http.Request) (*http.Response, error)
	DoWithRatingResponse(request *http.Request) (RatingResponse, error)
	GetBaseUrl() *url.URL
}

type FilterService interface {
	ApplyFiltersToRequest(request *http.Request, filters []model.Filter)
}

type SearchService interface {
	GetResults(ctx context.Context, item model.Item) ([]model.SearchResult, error)
}

type RatingPlatformService interface {
	GetRating(ctx context.Context, item model.Item) (model.Rating, error)
}

type LogoService interface {
	GetLogos(ctx context.Context, ratings []model.Rating, itemID string, dimensions model.LogoDimensions) ([]*model.Logo, error)
}
