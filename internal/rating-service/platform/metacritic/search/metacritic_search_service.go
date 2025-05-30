package metacritic

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
	metacritic "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/metacritic/model"
)

type MetacriticSearchService struct {
	client         rating.RatingClient
	filtersService rating.FilterService
	logger         *zap.Logger
}

func NewMetacriticSearchService(client rating.RatingClient, filtersService rating.FilterService, logger *zap.Logger) rating.SearchService {
	return &MetacriticSearchService{
		client:         client,
		filtersService: filtersService,
		logger:         logger,
	}
}

func (s *MetacriticSearchService) GetResults(ctx context.Context, item model.Item) ([]model.SearchResult, error) {
	var searchResults []model.SearchResult

	baseUrl := s.client.GetBaseUrl()

	// Construct the search URL with proper query parameters
	// We should use the original title to search for the item, because metacritic doesn't support language specific search, but fallback to the title if the original title is empty
	titleForSearch := item.OriginalTitle
	if titleForSearch == "" {
		titleForSearch = item.Title
	}

	searchPath := fmt.Sprintf("/finder/metacritic/search/%s/web", url.QueryEscape(strings.ToLower(titleForSearch)))
	endpoint := baseUrl.JoinPath(searchPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		s.logger.Error("unable to build request",
			zap.String("method", "GetResults"),
			zap.String("url", endpoint.String()),
			zap.Error(err),
		)
		return nil, err
	}

	response, err := s.client.DoWithRatingResponse(req)
	if err != nil {
		s.logger.Error("unable to perform request to Metacritic",
			zap.String("method", "GetResults"),
			zap.Error(err),
		)
		return searchResults, err
	}

	results, ok := response.(*metacritic.Response)
	if !ok {
		s.logger.Error("unable to cast response to Metacritic Response",
			zap.String("method", "GetResults"),
		)
		return nil, fmt.Errorf("invalid response type")
	}

	// Filter results by year
	items := lo.Filter(results.Data.Items, func(result metacritic.Entry, _ int) bool {
		return result.PremiereYear == item.Year
	})

	return s.convertMetacriticResultsToSearchResults(items), nil
}

func (s *MetacriticSearchService) convertMetacriticResultsToSearchResults(results []metacritic.Entry) []model.SearchResult {
	return lo.Map(results, func(result metacritic.Entry, _ int) model.SearchResult {
		return model.SearchResult{
			ID:    result.ID,
			Title: result.Title,
			Vote:  float64(result.CriticScore.Score),
		}
	})
}
