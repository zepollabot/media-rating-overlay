package tmdb

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	ratingclientmock "github.com/zepollabot/media-rating-overlay/internal/rating-service/mocks"
	tmdbmodel "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/model"
)

type TMDBSearchServiceTestSuite struct {
	suite.Suite
	mockClient         *ratingclientmock.RatingClient
	mockFiltersService *ratingclientmock.FilterService
	logger             *zap.Logger
	service            *TMDBSearchService
	ctx                context.Context
}

func TestTMDBSearchServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TMDBSearchServiceTestSuite))
}

func (s *TMDBSearchServiceTestSuite) SetupTest() {
	s.mockClient = ratingclientmock.NewRatingClient(s.T())
	s.mockFiltersService = ratingclientmock.NewFilterService(s.T())
	s.logger = zap.NewNop() // Use a no-op logger for tests
	s.service = NewTMDBSearchService(s.mockClient, s.mockFiltersService, s.logger).(*TMDBSearchService)
	s.ctx = context.Background()
}

func (s *TMDBSearchServiceTestSuite) TearDownTest() {
	s.mockClient.AssertExpectations(s.T())
	s.mockFiltersService.AssertExpectations(s.T())
}

func (s *TMDBSearchServiceTestSuite) TestGetResults_Success() {
	// Arrange
	item := model.Item{Title: "Inception", Year: 2010}
	expectedBaseURL, _ := url.Parse("http://test.com")
	expectedEndpoint := "http://test.com/3/search/movie"

	tmdbResults := []tmdbmodel.Entry{
		{ID: 1, Title: "Inception", Vote: 8.8},
		{ID: 2, Title: "Inception: The IMAX Experience", Vote: 8.5},
	}
	expectedSearchResults := []model.SearchResult{
		{ID: 1, Title: "Inception", Vote: 8.8},
		{ID: 2, Title: "Inception: The IMAX Experience", Vote: 8.5},
	}

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.On("ApplyFiltersToRequest", mock.AnythingOfType("*http.Request"), mock.MatchedBy(func(filters []model.Filter) bool {
		return len(filters) == 2 &&
			filters[0].Name == "query" && filters[0].Value == item.Title &&
			filters[1].Name == "year" && filters[1].Value == strconv.Itoa(item.Year)
	})).Return()
	s.mockClient.On("DoWithRatingResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == expectedEndpoint && req.Method == http.MethodGet
	})).Return(&tmdbmodel.Response{Results: tmdbResults}, nil)

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.NoError(err)
	s.Equal(expectedSearchResults, results)
}

func (s *TMDBSearchServiceTestSuite) TestGetResults_HttpNewRequestWithContextError() {
	// Arrange
	item := model.Item{Title: "Error Movie", Year: 2020}
	// Construct a base URL that will cause http.NewRequestWithContext to fail
	// due to an invalid character in the host part when endpoint.String() is called.
	malformedBaseURL := &url.URL{Scheme: "http", Host: "local\x00host", Path: "/"}

	s.mockClient.On("GetBaseUrl").Return(malformedBaseURL)
	// s.mockFiltersService.ApplyFiltersToRequest should not be called
	// s.mockClient.DoWithRatingResponse should not be called

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.Error(err)
	s.Nil(results)
	// Check for the specific error from url.Parse, which NewRequestWithContext uses
	urlErr, ok := err.(*url.Error)
	s.True(ok, "error should be of type *url.Error")
	s.Equal("parse", urlErr.Op)
	s.Contains(urlErr.Err.Error(), "invalid URL")
}

func (s *TMDBSearchServiceTestSuite) TestGetResults_ClientDoError() {
	// Arrange
	item := model.Item{Title: "Error Movie", Year: 2020}
	expectedBaseURL, _ := url.Parse("http://test.com")
	clientError := errors.New("network error")

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.On("ApplyFiltersToRequest", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("[]model.Filter")).Return()
	s.mockClient.On("DoWithRatingResponse", mock.AnythingOfType("*http.Request")).Return(nil, clientError)

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.Error(err)
	s.Nil(results)
	s.Equal(clientError, err)
}

func (s *TMDBSearchServiceTestSuite) TestGetResults_InvalidResponseType() {
	// Arrange
	item := model.Item{Title: "Invalid Response Movie", Year: 2021}
	expectedBaseURL, _ := url.Parse("http://test.com")

	// Return a response type that is not *tmdbmodel.Response
	invalidResponse := struct{ Message string }{"I am not a TMDB response"}

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.On("ApplyFiltersToRequest", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("[]model.Filter")).Return()
	s.mockClient.On("DoWithRatingResponse", mock.AnythingOfType("*http.Request")).Return(invalidResponse, nil)

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.Error(err)
	s.Nil(results)
	s.Contains(err.Error(), "invalid response type")
}

func (s *TMDBSearchServiceTestSuite) TestGetResults_EmptyResults() {
	// Arrange
	item := model.Item{Title: "No Results Movie", Year: 2022}
	expectedBaseURL, _ := url.Parse("http://test.com")

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.On("ApplyFiltersToRequest", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("[]model.Filter")).Return()
	s.mockClient.On("DoWithRatingResponse", mock.AnythingOfType("*http.Request")).Return(&tmdbmodel.Response{Results: []tmdbmodel.Entry{}}, nil)

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.NoError(err)
	s.Empty(results)
}

func (s *TMDBSearchServiceTestSuite) TestGetResults_FilterApplication() {
	// Arrange
	item := model.Item{Title: "Filtered Movie", Year: 2023}
	expectedBaseURL, _ := url.Parse("http://test.com")
	expectedFilters := []model.Filter{
		{Name: "query", Value: item.Title},
		{Name: "year", Value: strconv.Itoa(item.Year)},
	}

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.On("ApplyFiltersToRequest", mock.AnythingOfType("*http.Request"), expectedFilters).Return().Once() // Ensure it's called once with correct filters
	s.mockClient.On("DoWithRatingResponse", mock.AnythingOfType("*http.Request")).Return(&tmdbmodel.Response{Results: []tmdbmodel.Entry{}}, nil)

	// Act
	_, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.NoError(err)
	// AssertExpectations in TearDownTest will verify the mock calls.
}

func (s *TMDBSearchServiceTestSuite) TestConvertTMDBResultsToSearchResults() {
	// Arrange
	tmdbEntries := []tmdbmodel.Entry{
		{ID: 1, Title: "Movie A", OriginalTitle: "Original A", Vote: 7.5},
		{ID: 2, Title: "Movie B", OriginalTitle: "Original B", Vote: 8.0},
	}
	expectedSearchResults := []model.SearchResult{
		{ID: 1, Title: "Movie A", Vote: 7.5},
		{ID: 2, Title: "Movie B", Vote: 8.0},
	}

	// Act
	actualSearchResults := s.service.convertTMDBResultsToSearchResults(tmdbEntries)

	// Assert
	s.Equal(expectedSearchResults, actualSearchResults)
}

func (s *TMDBSearchServiceTestSuite) TestConvertTMDBResultsToSearchResults_Empty() {
	// Arrange
	tmdbEntries := []tmdbmodel.Entry{}
	expectedSearchResults := []model.SearchResult{}

	// Act
	actualSearchResults := s.service.convertTMDBResultsToSearchResults(tmdbEntries)

	// Assert
	s.Equal(expectedSearchResults, actualSearchResults)
}
