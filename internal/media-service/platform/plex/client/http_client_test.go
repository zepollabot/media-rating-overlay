package plex

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	common "github.com/zepollabot/media-rating-overlay/internal/httpclient"
	httpclientmocks "github.com/zepollabot/media-rating-overlay/internal/httpclient/mocks"
)

// PlexHTTPClientTestSuite defines the test suite for PlexHTTPClient
type PlexHTTPClientTestSuite struct {
	suite.Suite
	mockCommonClient *httpclientmocks.HTTPClient
	plexHTTPClient   *PlexHTTPClient
}

// SetupTest sets up the test environment
func (s *PlexHTTPClientTestSuite) SetupTest() {
	s.mockCommonClient = new(httpclientmocks.HTTPClient)
	// In a real scenario, NewPlexHTTPClient would create a real common.NewRetryClient.
	// For testing PlexHTTPClient itself, we'll inject our mock into the created client.
	s.plexHTTPClient = NewPlexHTTPClient(5*time.Second, 3)
	s.plexHTTPClient.client = s.mockCommonClient // Inject mock
}

// TearDownTest asserts mock expectations
func (s *PlexHTTPClientTestSuite) TearDownTest() {
	s.mockCommonClient.AssertExpectations(s.T())
}

// TestPlexHTTPClientTestSuite runs the test suite
func TestPlexHTTPClientTestSuite(t *testing.T) {
	suite.Run(t, new(PlexHTTPClientTestSuite))
}

func (s *PlexHTTPClientTestSuite) TestNewPlexHTTPClient() {
	// Arrange
	timeout := 10 * time.Second
	maxRetries := 5

	// Act
	client := NewPlexHTTPClient(timeout, maxRetries)

	// Assert
	s.NotNil(client)
	s.NotNil(client.client, "Internal common client should be initialized")
}

func (s *PlexHTTPClientTestSuite) TestPlexHTTPClient_Do_Success() {
	// Arrange
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	expectedResp := &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}

	s.mockCommonClient.On("Do", req).Return(expectedResp, nil).Once()

	// Act
	resp, err := s.plexHTTPClient.Do(req)

	// Assert
	s.NoError(err)
	s.Equal(expectedResp, resp)
}

func (s *PlexHTTPClientTestSuite) TestPlexHTTPClient_Do_Error() {
	// Arrange
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	expectedErr := errors.New("do error")

	s.mockCommonClient.On("Do", req).Return(nil, expectedErr).Once()

	// Act
	resp, err := s.plexHTTPClient.Do(req)

	// Assert
	s.Error(err)
	s.Equal(expectedErr, err)
	s.Nil(resp)
}

// Test to ensure that if the underlying client is a real retry client, it's being used.
// This is more of an integration test snippet that shows how NewPlexHTTPClient is intended to work.
func (s *PlexHTTPClientTestSuite) TestNewPlexHTTPClient_IntegrationWithRealRetryClient() {
	// Arrange
	timeout := 1 * time.Millisecond // Short timeout to make retries happen fast
	maxRetries := 2

	// Act
	plexClientWithRealRetry := NewPlexHTTPClient(timeout, maxRetries)

	// Assert
	s.NotNil(plexClientWithRealRetry)
	s.NotNil(plexClientWithRealRetry.client)

	var httpClientType common.HTTPClient
	s.Assert().Implements(&httpClientType, plexClientWithRealRetry.client, "Client should implement common.HTTPClient")
}
