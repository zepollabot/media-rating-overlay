package tmdb

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
)

type TMDBFilterService struct {
	logger *zap.Logger
}

func NewTMDBFilterService(logger *zap.Logger) rating.FilterService {
	return &TMDBFilterService{
		logger: logger,
	}
}

func (s *TMDBFilterService) ApplyFiltersToRequest(request *http.Request, filters []model.Filter) {
	q := request.URL.Query()

	for _, filter := range filters {
		if filter.Name != "" && filter.Value != "" {
			q.Add(filter.Name, filter.Value)
		}
	}

	request.URL.RawQuery = q.Encode()
}
