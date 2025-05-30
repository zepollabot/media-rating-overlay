package metacritic

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
)

type MetacriticFilterService struct {
	logger *zap.Logger
}

func NewMetacriticFilterService(logger *zap.Logger) rating.FilterService {
	return &MetacriticFilterService{
		logger: logger,
	}
}

func (s *MetacriticFilterService) ApplyFiltersToRequest(request *http.Request, filters []model.Filter) {
	// Do nothing, Metacritic doesn't support filters on the request
}
