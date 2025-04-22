package tmdb_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
	tmdbFilter "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/filter"
)

type TMDBFilterServiceTestSuite struct {
	suite.Suite
	logger  *zap.Logger
	service rating.FilterService
}

func (s *TMDBFilterServiceTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.service = tmdbFilter.NewTMDBFilterService(s.logger)
}

func TestTMDBFilterServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TMDBFilterServiceTestSuite))
}

func (s *TMDBFilterServiceTestSuite) TestNewTMDBFilterService() {
	// Arrange
	logger := zap.NewNop()

	// Act
	service := tmdbFilter.NewTMDBFilterService(logger)

	// Assert
	s.NotNil(service)
	_, ok := service.(*tmdbFilter.TMDBFilterService)
	s.True(ok, "Service should be of type *tmdbFilter.TMDBFilterService")
}

func (s *TMDBFilterServiceTestSuite) TestApplyFiltersToRequest_NoFilters() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api")
	req := &http.Request{URL: reqURL}
	var filters []model.Filter

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	s.Equal("", req.URL.RawQuery, "RawQuery should be empty when no filters are applied")
}

func (s *TMDBFilterServiceTestSuite) TestApplyFiltersToRequest_SingleValidFilter() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api")
	req := &http.Request{URL: reqURL}
	filters := []model.Filter{
		{Name: "year", Value: "2023"},
	}

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	parsedQuery, _ := url.ParseQuery(req.URL.RawQuery)
	s.Equal("2023", parsedQuery.Get("year"))
}

func (s *TMDBFilterServiceTestSuite) TestApplyFiltersToRequest_MultipleValidFilters() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api")
	req := &http.Request{URL: reqURL}
	filters := []model.Filter{
		{Name: "year", Value: "2023"},
		{Name: "genre", Value: "action"},
	}

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	parsedQuery, _ := url.ParseQuery(req.URL.RawQuery)
	s.Equal("2023", parsedQuery.Get("year"))
	s.Equal("action", parsedQuery.Get("genre"))
}

func (s *TMDBFilterServiceTestSuite) TestApplyFiltersToRequest_WithEmptyNameFilter() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api")
	req := &http.Request{URL: reqURL}
	filters := []model.Filter{
		{Name: "", Value: "2023"},
		{Name: "genre", Value: "action"},
	}

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	parsedQuery, _ := url.ParseQuery(req.URL.RawQuery)
	s.Equal("action", parsedQuery.Get("genre"))
	s.Empty(parsedQuery.Get(""), "Filter with empty name should not be added")
	s.Len(parsedQuery, 1)
}

func (s *TMDBFilterServiceTestSuite) TestApplyFiltersToRequest_WithEmptyValueFilter() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api")
	req := &http.Request{URL: reqURL}
	filters := []model.Filter{
		{Name: "year", Value: ""},
		{Name: "genre", Value: "action"},
	}

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	parsedQuery, _ := url.ParseQuery(req.URL.RawQuery)
	s.Equal("action", parsedQuery.Get("genre"))
	s.Empty(parsedQuery.Get("year"), "Filter with empty value should not be added")
	s.Len(parsedQuery, 1)
}

func (s *TMDBFilterServiceTestSuite) TestApplyFiltersToRequest_WithExistingQueryParameters() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api?existing_param=true")
	req := &http.Request{URL: reqURL}
	filters := []model.Filter{
		{Name: "year", Value: "2023"},
	}

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	parsedQuery, _ := url.ParseQuery(req.URL.RawQuery)
	s.Equal("true", parsedQuery.Get("existing_param"))
	s.Equal("2023", parsedQuery.Get("year"))
	s.Len(parsedQuery, 2)
}

func (s *TMDBFilterServiceTestSuite) TestApplyFiltersToRequest_FilterWithSpecialCharacters() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api")
	req := &http.Request{URL: reqURL}
	filters := []model.Filter{
		{Name: "queryParam", Value: "value with spaces & special=chars"},
	}

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	parsedQuery, _ := url.ParseQuery(req.URL.RawQuery)
	s.Equal("value with spaces & special=chars", parsedQuery.Get("queryParam"))
}
