package tmdb

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
	tmdb "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/model"
)

type TMDBSearchService struct {
	client         rating.RatingClient
	filtersService rating.FilterService
	logger         *zap.Logger
}

func NewTMDBSearchService(client rating.RatingClient, filtersService rating.FilterService, logger *zap.Logger) rating.SearchService {
	return &TMDBSearchService{
		client:         client,
		filtersService: filtersService,
		logger:         logger,
	}
}

func (s *TMDBSearchService) GetResults(ctx context.Context, item model.Item) ([]model.SearchResult, error) {
	var searchResults []model.SearchResult

	baseUrl := s.client.GetBaseUrl()
	endpoint := baseUrl.JoinPath("/3/search/movie")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		s.logger.Error("unable to build request",
			zap.String("method", "GetEntries"),
			zap.String("url", endpoint.String()),
			zap.Error(err),
		)
		return nil, err
	}

	filters := []model.Filter{
		{
			Name:  "query",
			Value: item.Title,
		},
		{
			Name:  "year",
			Value: strconv.Itoa(item.Year),
		},
	}

	s.filtersService.ApplyFiltersToRequest(req, filters)

	response, err := s.client.DoWithRatingResponse(req)

	if err != nil {
		s.logger.Error("unable to perform request to TMDB Client",
			zap.String("method", "GetEntries"),
			zap.Error(err),
		)
		return searchResults, err
	}

	results, ok := response.(*tmdb.Response)
	if !ok {
		s.logger.Error("unable to cast response to TMDB Response",
			zap.String("method", "GetEntries"),
		)
		return nil, fmt.Errorf("invalid response type")
	}

	return s.convertTMDBResultsToSearchResults(results.Results), nil
}

func (s *TMDBSearchService) convertTMDBResultsToSearchResults(results []tmdb.Entry) []model.SearchResult {
	return lo.Map(results, func(result tmdb.Entry, _ int) model.SearchResult {
		return model.SearchResult{
			ID:    result.ID,
			Title: result.Title,
			Vote:  result.Vote,
		}
	})
}
