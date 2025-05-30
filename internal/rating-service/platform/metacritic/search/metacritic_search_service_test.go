package metacritic

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	ratingclientmock "github.com/zepollabot/media-rating-overlay/internal/rating-service/mocks"
	metacriticmodel "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/metacritic/model"
)

type MetacriticSearchServiceTestSuite struct {
	suite.Suite
	mockClient         *ratingclientmock.RatingClient
	mockFiltersService *ratingclientmock.FilterService
	logger             *zap.Logger
	service            *MetacriticSearchService
	ctx                context.Context
}

func TestMetacriticSearchServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MetacriticSearchServiceTestSuite))
}

func (s *MetacriticSearchServiceTestSuite) SetupTest() {
	s.mockClient = ratingclientmock.NewRatingClient(s.T())
	s.mockFiltersService = ratingclientmock.NewFilterService(s.T())
	s.logger = zap.NewNop() // Use a no-op logger for tests
	s.service = NewMetacriticSearchService(s.mockClient, s.mockFiltersService, s.logger).(*MetacriticSearchService)
	s.ctx = context.Background()
}

func (s *MetacriticSearchServiceTestSuite) TearDownTest() {
	s.mockClient.AssertExpectations(s.T())
	s.mockFiltersService.AssertExpectations(s.T())
}

func (s *MetacriticSearchServiceTestSuite) TestGetResults_Success() {
	// Arrange
	item := model.Item{Title: "Inception", Year: 2010}
	expectedBaseURL, _ := url.Parse("http://test.com")

	metacriticResults := []metacriticmodel.Entry{
		{
			ID:    1,
			Title: "Inception",
			CriticScore: struct {
				URL   string `json:"url"`
				Score int    `json:"score"`
			}{Score: 75},
			PremiereYear: 2010,
		},
		{
			ID:    2,
			Title: "Inception: The IMAX Experience",
			CriticScore: struct {
				URL   string `json:"url"`
				Score int    `json:"score"`
			}{Score: 80},
			PremiereYear: 2010,
		},
	}
	expectedSearchResults := []model.SearchResult{
		{ID: 1, Title: "Inception", Vote: 75},
		{ID: 2, Title: "Inception: The IMAX Experience", Vote: 80},
	}

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.AssertNotCalled(s.T(), "ApplyFiltersToRequest")
	s.mockClient.On("DoWithRatingResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.Path == "/finder/metacritic/search/inception/web" && req.Method == http.MethodGet
	})).Return(&metacriticmodel.Response{Data: metacriticmodel.Data{Items: metacriticResults}}, nil)

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.NoError(err)
	s.Equal(expectedSearchResults, results)
}

func (s *MetacriticSearchServiceTestSuite) TestGetResults_HttpNewRequestWithContextError() {
	// Arrange
	item := model.Item{Title: "Error Movie", Year: 2020}
	// Construct a base URL that will cause http.NewRequestWithContext to fail
	// due to an invalid character in the host part when endpoint.String() is called.
	malformedBaseURL := &url.URL{Scheme: "http", Host: "local\x00host", Path: "/"}

	s.mockClient.On("GetBaseUrl").Return(malformedBaseURL)
	s.mockFiltersService.AssertNotCalled(s.T(), "ApplyFiltersToRequest")
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

func (s *MetacriticSearchServiceTestSuite) TestGetResults_ClientDoError() {
	// Arrange
	item := model.Item{Title: "Error Movie", Year: 2020}
	expectedBaseURL, _ := url.Parse("http://test.com")
	clientError := errors.New("network error")

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.AssertNotCalled(s.T(), "ApplyFiltersToRequest")
	s.mockClient.On("DoWithRatingResponse", mock.AnythingOfType("*http.Request")).Return(nil, clientError)

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.Error(err)
	s.Nil(results)
	s.Equal(clientError, err)
}

func (s *MetacriticSearchServiceTestSuite) TestGetResults_InvalidResponseType() {
	// Arrange
	item := model.Item{Title: "Invalid Response Movie", Year: 2021}
	expectedBaseURL, _ := url.Parse("http://test.com")

	// Return a response type that is not *tmdbmodel.Response
	invalidResponse := struct{ Message string }{"I am not a Metacritic response"}

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.AssertNotCalled(s.T(), "ApplyFiltersToRequest")
	s.mockClient.On("DoWithRatingResponse", mock.AnythingOfType("*http.Request")).Return(invalidResponse, nil)

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.Error(err)
	s.Nil(results)
	s.Contains(err.Error(), "invalid response type")
}

func (s *MetacriticSearchServiceTestSuite) TestGetResults_EmptyResults() {
	// Arrange
	item := model.Item{Title: "No Results Movie", Year: 2022}
	expectedBaseURL, _ := url.Parse("http://test.com")

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.AssertNotCalled(s.T(), "ApplyFiltersToRequest")
	s.mockClient.On("DoWithRatingResponse", mock.AnythingOfType("*http.Request")).Return(&metacriticmodel.Response{Data: metacriticmodel.Data{Items: []metacriticmodel.Entry{}}}, nil)

	// Act
	results, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.NoError(err)
	s.Empty(results)
}

func (s *MetacriticSearchServiceTestSuite) TestGetResults_FilterApplication() {
	// Arrange
	item := model.Item{Title: "Filtered Movie", Year: 2023}
	expectedBaseURL, _ := url.Parse("http://test.com")

	s.mockClient.On("GetBaseUrl").Return(expectedBaseURL)
	s.mockFiltersService.AssertNotCalled(s.T(), "ApplyFiltersToRequest")
	s.mockClient.On("DoWithRatingResponse", mock.AnythingOfType("*http.Request")).Return(&metacriticmodel.Response{Data: metacriticmodel.Data{Items: []metacriticmodel.Entry{}}}, nil)

	// Act
	_, err := s.service.GetResults(s.ctx, item)

	// Assert
	s.NoError(err)
	// AssertExpectations in TearDownTest will verify the mock calls.
}

func (s *MetacriticSearchServiceTestSuite) TestConvertMetacriticResultsToSearchResults() {
	// Arrange
	metacriticEntries := []metacriticmodel.Entry{
		{
			ID:    1,
			Title: "Movie A",
			CriticScore: struct {
				URL   string `json:"url"`
				Score int    `json:"score"`
			}{Score: 75},
		},
		{
			ID:    2,
			Title: "Movie B",
			CriticScore: struct {
				URL   string `json:"url"`
				Score int    `json:"score"`
			}{Score: 80},
		},
	}
	expectedSearchResults := []model.SearchResult{
		{ID: 1, Title: "Movie A", Vote: 75},
		{ID: 2, Title: "Movie B", Vote: 80},
	}

	// Act
	actualSearchResults := s.service.convertMetacriticResultsToSearchResults(metacriticEntries)

	// Assert
	s.Equal(expectedSearchResults, actualSearchResults)
}

func (s *MetacriticSearchServiceTestSuite) TestConvertMetacriticResultsToSearchResults_Empty() {
	// Arrange
	metacriticEntries := []metacriticmodel.Entry{}
	expectedSearchResults := []model.SearchResult{}

	// Act
	actualSearchResults := s.service.convertMetacriticResultsToSearchResults(metacriticEntries)

	// Assert
	s.Equal(expectedSearchResults, actualSearchResults)
}
