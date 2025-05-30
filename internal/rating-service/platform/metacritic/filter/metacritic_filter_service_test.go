package metacritic_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
	metacriticFilter "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/metacritic/filter"
)

type MetacriticFilterServiceTestSuite struct {
	suite.Suite
	logger  *zap.Logger
	service rating.FilterService
}

func (s *MetacriticFilterServiceTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.service = metacriticFilter.NewMetacriticFilterService(s.logger)
}

func TestMetacriticFilterServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MetacriticFilterServiceTestSuite))
}

func (s *MetacriticFilterServiceTestSuite) TestNewMetacriticFilterService() {
	// Arrange
	logger := zap.NewNop()

	// Act
	service := metacriticFilter.NewMetacriticFilterService(logger)

	// Assert
	s.NotNil(service)
	_, ok := service.(*metacriticFilter.MetacriticFilterService)
	s.True(ok, "Service should be of type *metacriticFilter.MetacriticFilterService")
}

func (s *MetacriticFilterServiceTestSuite) TestApplyFiltersToRequest_NoFilters() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api")
	req := &http.Request{URL: reqURL}
	var filters []model.Filter

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	s.Equal("", req.URL.RawQuery, "RawQuery should be empty when no filters are applied")
}

func (s *MetacriticFilterServiceTestSuite) TestApplyFiltersToRequest_ShouldNotModifyRequest() {
	// Arrange
	reqURL, _ := url.Parse("http://example.com/api?mcoTypeId=2&sortDirection=DESC&limit=24&offset=0")
	req := &http.Request{URL: reqURL}
	var filters []model.Filter

	// Act
	s.service.ApplyFiltersToRequest(req, filters)

	// Assert
	s.Equal("http://example.com/api?mcoTypeId=2&sortDirection=DESC&limit=24&offset=0", req.URL.String(), "Request URL should not be modified")
}
