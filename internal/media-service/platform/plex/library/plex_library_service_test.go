package plex_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	mediaClientMock "github.com/zepollabot/media-rating-overlay/internal/media-service/mocks"
	plexlibrary "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/library"
	plexmodel "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/model"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type PlexLibraryServiceTestSuite struct {
	suite.Suite
	mockClient *mediaClientMock.MediaClient
	service    plexlibrary.LibraryService
	logger     *zap.Logger
}

func (s *PlexLibraryServiceTestSuite) SetupTest() {
	s.mockClient = mediaClientMock.NewMediaClient(s.T())
	s.logger = zap.NewNop() // Use a no-op logger for tests
	s.service = plexlibrary.NewPlexLibraryService(s.mockClient, s.logger)
}

func (s *PlexLibraryServiceTestSuite) TearDownTest() {
	s.mockClient.AssertExpectations(s.T())
}

func TestPlexLibraryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PlexLibraryServiceTestSuite))
}

func (s *PlexLibraryServiceTestSuite) TestGetLibraries_Success() {
	// Arrange
	ctx := context.Background()
	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := "http://localhost:32400/library/sections"

	plexLibs := []plexmodel.Library{
		{ID: "1", Title: "Movies", Type: "movie", Language: "en"},
		{ID: "2", Title: "TV Shows", Type: "show", Language: "en"},
	}
	plexResponse := &plexmodel.Response{
		MediaContainer: plexmodel.MediaContainer{
			Libraries: plexLibs,
		},
	}

	expectedLibs := []model.Library{
		{ID: "1", Name: "Movies", Type: "movie", Language: "en"},
		{ID: "2", Name: "TV Shows", Type: "show", Language: "en"},
	}

	s.mockClient.On("GetBaseUrl").Return(baseURL)
	s.mockClient.On("DoWithMediaResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.String() == expectedEndpoint
	})).Return(plexResponse, nil)

	// Act
	libs, err := s.service.GetLibraries(ctx)

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedLibs, libs)
}

func (s *PlexLibraryServiceTestSuite) TestGetLibraries_HttpRequestError() {
	// Arrange
	ctx := context.Background()
	baseURL, _ := url.Parse("http://localhost:32400")
	// Forcing an error in NewRequestWithContext by providing an invalid method
	// This is a bit artificial as the service constructs it correctly.
	// A more realistic test would be to mock http.NewRequestWithContext if possible,
	// or ensure the service handles errors from it, though it currently doesn't directly expose such a failure path
	// given the current implementation. For now, we'll test error from client.DoWithMediaResponse.

	s.mockClient.On("GetBaseUrl").Return(baseURL)
	// We can't easily inject an error into http.NewRequestWithContext from the test
	// as it's called internally. Instead, we'll test an error from DoWithMediaResponse

	// Let's test the path where client.DoWithMediaResponse returns an error
	expectedErr := errors.New("client error")
	s.mockClient.On("DoWithMediaResponse", mock.AnythingOfType("*http.Request")).Return(nil, expectedErr)

	// Act
	libs, err := s.service.GetLibraries(ctx)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), libs)
	assert.EqualError(s.T(), err, expectedErr.Error())
}

func (s *PlexLibraryServiceTestSuite) TestGetLibraries_PlexClientError() {
	// Arrange
	ctx := context.Background()
	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := "http://localhost:32400/library/sections"
	clientError := errors.New("plex client error")

	s.mockClient.On("GetBaseUrl").Return(baseURL)
	s.mockClient.On("DoWithMediaResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.String() == expectedEndpoint
	})).Return(nil, clientError)

	// Act
	libs, err := s.service.GetLibraries(ctx)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), libs)
	assert.Equal(s.T(), clientError, err)
}

func (s *PlexLibraryServiceTestSuite) TestGetLibraries_InvalidResponseType() {
	// Arrange
	ctx := context.Background()
	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := "http://localhost:32400/library/sections"

	// Return a different type instead of *plex.Response
	invalidResponse := &mediaClientMock.MediaResponse{} // Using a mock type or any other non-plex.Response type

	s.mockClient.On("GetBaseUrl").Return(baseURL)
	s.mockClient.On("DoWithMediaResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.String() == expectedEndpoint
	})).Return(invalidResponse, nil)

	// Act
	libs, err := s.service.GetLibraries(ctx)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), libs)
	assert.EqualError(s.T(), err, "invalid response type")
}

func (s *PlexLibraryServiceTestSuite) TestGetLibraries_EmptyLibraries() {
	// Arrange
	ctx := context.Background()
	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := "http://localhost:32400/library/sections"

	plexResponse := &plexmodel.Response{
		MediaContainer: plexmodel.MediaContainer{
			Libraries: []plexmodel.Library{}, // Empty slice
		},
	}

	s.mockClient.On("GetBaseUrl").Return(baseURL)
	s.mockClient.On("DoWithMediaResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.String() == expectedEndpoint
	})).Return(plexResponse, nil)

	// Act
	libs, err := s.service.GetLibraries(ctx)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), libs)
	assert.Empty(s.T(), libs)
}

func (s *PlexLibraryServiceTestSuite) TestRefreshLibrary_Success_NoForce() {
	// Arrange
	ctx := context.Background()
	libraryID := "123"
	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := fmt.Sprintf("http://localhost:32400/library/sections/%s/refresh", libraryID)

	s.mockClient.On("GetBaseUrl").Return(baseURL)
	s.mockClient.On("DoWithResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet &&
			req.URL.String() == expectedEndpoint &&
			req.URL.Query().Get("force") == ""
	})).Return(&http.Response{StatusCode: http.StatusOK}, nil)

	// Act
	err := s.service.RefreshLibrary(ctx, libraryID, false)

	// Assert
	assert.NoError(s.T(), err)
}

func (s *PlexLibraryServiceTestSuite) TestRefreshLibrary_Success_WithForce() {
	// Arrange
	ctx := context.Background()
	libraryID := "123"
	baseURL, _ := url.Parse("http://localhost:32400")
	// The expected endpoint for the request itself, RawQuery will be checked separately or via MatchedBy
	baseEndpointPath := fmt.Sprintf("/library/sections/%s/refresh", libraryID)

	s.mockClient.On("GetBaseUrl").Return(baseURL)
	s.mockClient.On("DoWithResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet &&
			req.URL.Path == baseEndpointPath && // Check path separately
			req.URL.RawQuery == "force=1" // Check RawQuery
	})).Return(&http.Response{StatusCode: http.StatusOK}, nil)

	// Act
	err := s.service.RefreshLibrary(ctx, libraryID, true)

	// Assert
	assert.NoError(s.T(), err)
}

func (s *PlexLibraryServiceTestSuite) TestRefreshLibrary_PlexClientError() {
	// Arrange
	ctx := context.Background()
	libraryID := "456"
	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := fmt.Sprintf("http://localhost:32400/library/sections/%s/refresh", libraryID)
	clientError := errors.New("plex client error on refresh")

	s.mockClient.On("GetBaseUrl").Return(baseURL)
	s.mockClient.On("DoWithResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.String() == expectedEndpoint
	})).Return(nil, clientError)

	// Act
	err := s.service.RefreshLibrary(ctx, libraryID, false)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), clientError, err)
}

func (s *PlexLibraryServiceTestSuite) TestGetLibraries_NewRequestCreationError() {
	// Arrange
	ctx := context.Background()
	baseURL, _ := url.Parse("http://localhost:32400")
	expectedErr := errors.New("mocked NewRequestWithContext error")

	serviceInstance, ok := s.service.(*plexlibrary.PlexLibraryService)
	assert.True(s.T(), ok, "Service is not of type *plexlibrary.PlexLibraryService")

	originalNewRequestFunc := serviceInstance.NewRequestFunc
	serviceInstance.NewRequestFunc = func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
		return nil, expectedErr
	}
	defer func() { serviceInstance.NewRequestFunc = originalNewRequestFunc }()

	s.mockClient.On("GetBaseUrl").Return(baseURL)

	// Act
	libs, err := serviceInstance.GetLibraries(ctx)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), libs)
	assert.Equal(s.T(), expectedErr, err)
}

func (s *PlexLibraryServiceTestSuite) TestRefreshLibrary_NewRequestCreationError() {
	// Arrange
	ctx := context.Background()
	libraryID := "789"
	baseURL, _ := url.Parse("http://localhost:32400")
	expectedErr := errors.New("mocked NewRequestWithContext error for refresh")

	serviceInstance, ok := s.service.(*plexlibrary.PlexLibraryService)
	assert.True(s.T(), ok, "Service is not of type *plexlibrary.PlexLibraryService")

	originalNewRequestFunc := serviceInstance.NewRequestFunc
	serviceInstance.NewRequestFunc = func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
		return nil, expectedErr
	}
	defer func() { serviceInstance.NewRequestFunc = originalNewRequestFunc }()

	s.mockClient.On("GetBaseUrl").Return(baseURL)

	// Act
	err := serviceInstance.RefreshLibrary(ctx, libraryID, false)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
}

// Test for the scenario where http.NewRequestWithContext fails.
// This is a bit tricky to test directly without refactoring or using an http client mock
// that allows injecting errors at the request creation stage.
// For now, we'll focus on errors from client.DoWithResponse for RefreshLibrary.
// The internal error logging within RefreshLibrary for NewRequestWithContext
// would be the primary indicator of such a failure in a real scenario.
// A dedicated test TestRefreshLibrary_HttpRequestError would be similar to TestGetLibraries_HttpRequestError
// if we could reliably trigger the error in NewRequestWithContext for the RefreshLibrary call.
// Given current structure, error from DoWithResponse covers client communication issues.
