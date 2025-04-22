package plex

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	httpclientmocks "github.com/zepollabot/media-rating-overlay/internal/httpclient/mocks" // Alias for mocks
	plexModels "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/model"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
)

const (
	testPlexToken = "test-token"
	testPlexURL   = "http://localhost:32400"
	logFilePath   = "/tmp/test-plex-client.log"
)

type PlexClientTestSuite struct {
	suite.Suite
	mockHTTPClient        *httpclientmocks.ServiceHTTPClient
	mockGenericHTTPClient *httpclientmocks.HTTPClient // For SetHttpClient
	plexClient            *PlexClient
	logger                *zap.Logger
	plexConfig            *config.Plex
	httpClientConfig      *config.HTTPClient
}

func (s *PlexClientTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.mockHTTPClient = new(httpclientmocks.ServiceHTTPClient)
	s.mockGenericHTTPClient = new(httpclientmocks.HTTPClient)

	s.plexConfig = &config.Plex{
		Url:   testPlexURL,
		Token: testPlexToken,
	}
	s.httpClientConfig = &config.HTTPClient{
		Timeout:    10,
		MaxRetries: 3,
	}

	var err error
	s.plexClient, err = NewPlexClient(s.plexConfig, s.httpClientConfig, logFilePath, s.logger)
	s.Require().NoError(err)
	s.plexClient.httpClient = s.mockHTTPClient // Replace with mock
}

func (s *PlexClientTestSuite) TearDownTest() {
	s.mockHTTPClient.AssertExpectations(s.T())
	s.mockGenericHTTPClient.AssertExpectations(s.T())
}

func TestPlexClientTestSuite(t *testing.T) {
	suite.Run(t, new(PlexClientTestSuite))
}

func (s *PlexClientTestSuite) TestNewPlexClient_Success() {
	client, err := NewPlexClient(s.plexConfig, s.httpClientConfig, logFilePath, s.logger)
	s.NoError(err)
	s.NotNil(client)
	s.Equal(testPlexToken, client.token)
	s.Equal(testPlexURL, client.baseUrl.String())
	s.NotNil(client.httpClient) // httpClient is initialized internally, not the mock by default
}

func (s *PlexClientTestSuite) TestNewPlexClient_InvalidURL() {
	invalidConfig := &config.Plex{
		Url:   ":invalid-url", // Invalid URL
		Token: testPlexToken,
	}
	client, err := NewPlexClient(invalidConfig, s.httpClientConfig, logFilePath, s.logger)

	s.Error(err)
	s.Nil(client)
	s.Contains(err.Error(), "missing protocol scheme")
}

func (s *PlexClientTestSuite) TestNewPlexClient_EmptyURL() {
	invalidConfig := &config.Plex{
		Url:   "",
		Token: testPlexToken,
	}
	client, err := NewPlexClient(invalidConfig, s.httpClientConfig, logFilePath, s.logger)

	s.Error(err)
	s.Nil(client)
	s.Contains(err.Error(), "plex.url is required")
}

func (s *PlexClientTestSuite) TestPlexClient_DoWithResponse_Success() {
	// Arrange
	req, _ := http.NewRequest("GET", testPlexURL+"/status/sessions", nil)
	expectedResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("response body")),
	}
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
		request := args.Get(0).(*http.Request)
		s.Equal("application/json", request.Header.Get("Content-Type"))
		s.Equal("application/json", request.Header.Get("Accept"))
		s.Equal(testPlexToken, request.URL.Query().Get("X-Plex-Token"))
	}).Return(expectedResp, nil).Once()

	// Act
	resp, err := s.plexClient.DoWithResponse(req)

	// Assert
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(expectedResp, resp)
}

func (s *PlexClientTestSuite) TestPlexClient_DoWithResponse_HTTPClientError() {
	// Arrange
	req, _ := http.NewRequest("GET", testPlexURL+"/status/sessions", nil)
	expectedErr := errors.New("http client error")
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, expectedErr).Once()

	// Act
	resp, err := s.plexClient.DoWithResponse(req)

	// Assert
	s.Error(err)
	s.Nil(resp)
	s.Equal(expectedErr, err)
}

func (s *PlexClientTestSuite) TestPlexClient_DoWithMediaResponse_Success() {
	// Arrange
	req, _ := http.NewRequest("GET", testPlexURL+"/library/metadata/123", nil)
	mockResponseBody := `{
		"MediaContainer": {
			"size": 1,
			"Metadata": [{ "ratingKey": "123", "title": "Test Movie" }]
		}
	}`
	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(mockResponseBody)),
		Header:     make(http.Header),
	}
	httpResp.Header.Set("Content-Type", "application/json")

	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(httpResp, nil).Once()

	// Act
	mediaResp, err := s.plexClient.DoWithMediaResponse(req)

	// Assert
	s.NoError(err)
	s.NotNil(mediaResp)
	plexResp, ok := mediaResp.(*plexModels.Response)
	s.True(ok, "mediaResp should be of type *plexModels.Response")
	s.Equal(1, plexResp.MediaContainer.Size)
	s.Len(plexResp.MediaContainer.Entries, 1)
	s.Equal("123", plexResp.MediaContainer.Entries[0].ID)
}

func (s *PlexClientTestSuite) TestPlexClient_DoWithMediaResponse_DoWithResponseError() {
	// Arrange
	req, _ := http.NewRequest("GET", testPlexURL+"/library/metadata/123", nil)
	expectedErr := errors.New("do with response error")
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, expectedErr).Once()

	// Act
	mediaResp, err := s.plexClient.DoWithMediaResponse(req)

	// Assert
	s.Error(err)
	s.Nil(mediaResp)
	s.Equal(expectedErr, err)
}

func (s *PlexClientTestSuite) TestPlexClient_DoWithMediaResponse_Unauthorized() {
	// Arrange
	req, _ := http.NewRequest("GET", testPlexURL+"/library/metadata/123", nil)
	httpResp := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(strings.NewReader("")), // Body can be empty for unauthorized
		Header:     make(http.Header),
	}
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(httpResp, nil).Once()

	// Act
	mediaResp, err := s.plexClient.DoWithMediaResponse(req)

	// Assert
	s.Error(err)
	s.Nil(mediaResp)
	s.Equal(model.NotAuthorized, err.Error())
}

func (s *PlexClientTestSuite) TestPlexClient_DoWithMediaResponse_JSONDecodeError() {
	// Arrange
	req, _ := http.NewRequest("GET", testPlexURL+"/library/metadata/123", nil)
	malformedJSON := `{"MediaContainer": { "size": 1, "Metadata": [{ "ratingKey": "123", "title": "Test Movie" ]}}` // Missing closing } for Metadata array
	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(malformedJSON)),
		Header:     make(http.Header),
	}
	httpResp.Header.Set("Content-Type", "application/json")

	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(httpResp, nil).Once()

	// Act
	mediaResp, err := s.plexClient.DoWithMediaResponse(req)

	// Assert
	s.Error(err)
	s.Nil(mediaResp)
	s.Contains(err.Error(), "invalid character ']'")
}

func (s *PlexClientTestSuite) TestPlexClient_DoWithMediaResponse_BodyCloseError() {
	// Arrange
	req, _ := http.NewRequest("GET", testPlexURL+"/library/metadata/123", nil)
	mockResponseBody := `{
		"MediaContainer": {
			"size": 1,
			"Metadata": [{ "ratingKey": "123", "title": "Test Movie" }]
		}
	}`

	// Create a custom closer that returns an error on Close
	errOnCloseBody := &mockReadCloser{
		Reader:   strings.NewReader(mockResponseBody),
		closeErr: errors.New("failed to close body"),
	}

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       errOnCloseBody,
		Header:     make(http.Header),
	}
	httpResp.Header.Set("Content-Type", "application/json")

	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(httpResp, nil).Once()

	// For this test, we'll use a real logger to observe the error message
	var logBuffer bytes.Buffer
	writer := zapcore.AddSync(&logBuffer)
	encoderCfg := zap.NewProductionEncoderConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		writer,
		zap.InfoLevel,
	)
	testLogger := zap.New(core)
	s.plexClient.logger = testLogger // Temporarily replace logger

	// Act
	mediaResp, err := s.plexClient.DoWithMediaResponse(req)

	// Assert
	s.NoError(err) // The DoWithMediaResponse itself should not fail due to body close error
	s.NotNil(mediaResp)
	plexResp, ok := mediaResp.(*plexModels.Response)
	s.True(ok, "mediaResp should be of type *plexModels.Response")
	s.NotNil(plexResp)

	// Check content of response to ensure parsing happened before close error
	if plexResp != nil && len(plexResp.MediaContainer.Entries) > 0 {
		s.Equal("123", plexResp.MediaContainer.Entries[0].ID)
	}

	s.Contains(logBuffer.String(), "error closing response body")
	s.Contains(logBuffer.String(), "failed to close body")

	// Restore original logger
	s.plexClient.logger = zap.NewNop()
}

// mockReadCloser is a helper for testing body close errors
type mockReadCloser struct {
	io.Reader
	closeErr error
}

func (m *mockReadCloser) Close() error {
	return m.closeErr
}

func (s *PlexClientTestSuite) TestPlexClient_GetBaseUrl() {
	// Arrange
	expectedURL, _ := url.Parse(testPlexURL)

	// Act
	actualURL := s.plexClient.GetBaseUrl()

	// Assert
	s.Equal(expectedURL, actualURL)
}

func (s *PlexClientTestSuite) TestPlexClient_SetHttpClient() {
	// Arrange
	newMockClient := new(httpclientmocks.HTTPClient) // Use the generic HTTPClient mock

	// Act
	s.plexClient.SetHttpClient(newMockClient)

	// Assert
	s.Equal(newMockClient, s.plexClient.httpClient)

	// Ensure the original mock is not called if SetHttpClient changes the client
	// (This is more of a check for the test setup itself)
	req, _ := http.NewRequest("GET", testPlexURL+"/status/sessions", nil)
	expectedResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("response body")),
	}

	// Set up expectation on the new client
	newMockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(expectedResp, nil).Once()

	// Act with the new client
	_, err := s.plexClient.DoWithResponse(req)
	s.NoError(err)

	// Verify the new client was called
	newMockClient.AssertExpectations(s.T())
}

func (s *PlexClientTestSuite) TestPlexClient_setupRequest() {
	// Arrange
	reqURL := testPlexURL + "/some/path?existingParam=value"
	req, err := http.NewRequest("GET", reqURL, nil)
	s.Require().NoError(err)

	// Act
	err = s.plexClient.setupRequest(req)
	s.Require().NoError(err)

	// Assert
	s.Equal("application/json", req.Header.Get("Content-Type"))
	s.Equal("application/json", req.Header.Get("Accept"))

	query := req.URL.Query()
	s.Equal(testPlexToken, query.Get("X-Plex-Token"))
	s.Equal("value", query.Get("existingParam")) // Ensure existing params are preserved
}
