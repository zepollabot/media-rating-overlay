package tmdb_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	httpClientMocks "github.com/zepollabot/media-rating-overlay/internal/httpclient/mocks"
	"github.com/zepollabot/media-rating-overlay/internal/model"
	tmdbClient "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/client"
	tmdbModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/model"
)

type TMDBClientTestSuite struct {
	suite.Suite
	mockHTTPClient   *httpClientMocks.ServiceHTTPClient
	logger           *zap.Logger
	clientConfig     *config.TMDB
	httpClientConfig *config.HTTPClient
}

func (s *TMDBClientTestSuite) SetupTest() {
	s.mockHTTPClient = httpClientMocks.NewServiceHTTPClient(s.T())
	s.logger = zap.NewNop() // Use zap.NewExample() or zap.NewDevelopment() for verbose logs
	s.clientConfig = &config.TMDB{
		ApiKey:   "test-api-key",
		Language: "en",
		Region:   "US",
	}
	s.httpClientConfig = &config.HTTPClient{
		Timeout:    5, // 5 seconds
		MaxRetries: 3,
	}
}

func (s *TMDBClientTestSuite) TearDownTest() {
	s.mockHTTPClient.AssertExpectations(s.T())
}

func TestTMDBClientTestSuite(t *testing.T) {
	suite.Run(t, new(TMDBClientTestSuite))
}

// TestNewTMDBClient_Success tests the successful creation of a new TMDBClient
func (s *TMDBClientTestSuite) TestNewTMDBClient_Success() {
	// Arrange
	// SetupTest already prepares clientConfig, httpClientConfig, and logger

	// Act
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)

	// Assert
	s.NoError(err)
	s.NotNil(client)
	// Basic check, more detailed checks can be done on methods using these fields
	s.Equal("https", client.GetBaseUrl().Scheme)
	s.Equal("api.themoviedb.org", client.GetBaseUrl().Host)
}

func (s *TMDBClientTestSuite) TestGetBaseUrl() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)

	// Act
	baseUrl := client.GetBaseUrl()

	// Assert
	s.NotNil(baseUrl)
	s.Equal("https", baseUrl.Scheme)
	s.Equal("api.themoviedb.org", baseUrl.Host)
}

func (s *TMDBClientTestSuite) TestSetHttpClient() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)

	newMockHTTPClient := httpClientMocks.NewServiceHTTPClient(s.T()) // New mock for this test

	// Act
	client.SetHttpClient(newMockHTTPClient)

	// Assert
	// To verify, we'd ideally need a way to get the http client or see its effect.
	// For now, we'll assume the setter works if no panic and we can make a call.
	// This can be improved if there's a getter or if DoWithResponse uses it.
	// Let's try a simple Do call that we expect to use the new client.

	req, _ := http.NewRequest("GET", client.GetBaseUrl().String()+"/test", nil)
	newMockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{StatusCode: 200, Body: http.NoBody}, nil).Once()

	_, err = client.DoWithResponse(req)
	s.NoError(err)

	newMockHTTPClient.AssertExpectations(s.T()) // Assert on the new mock
}

func (s *TMDBClientTestSuite) TestDoWithResponse_Success() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)

	client.SetHttpClient(s.mockHTTPClient) // Use the suite's mock client

	mockRespBody := io.NopCloser(strings.NewReader(`{"message":"success"}`))
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       mockRespBody,
		Header:     make(http.Header),
	}

	var capturedRequest *http.Request
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).
		Run(func(args mock.Arguments) {
			capturedRequest = args.Get(0).(*http.Request)
		}).
		Return(mockResponse, nil).
		Once()

	reqURL := client.GetBaseUrl().String() + "/test/path"
	req, err := http.NewRequest("GET", reqURL, nil)
	s.Require().NoError(err)

	// Act
	resp, err := client.DoWithResponse(req)

	// Assert
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(http.StatusOK, resp.StatusCode)

	s.NotNil(capturedRequest)
	q := capturedRequest.URL.Query()
	s.Equal(s.clientConfig.ApiKey, q.Get("api_key"))
	s.Equal(s.clientConfig.Language, q.Get("language"))
	s.Equal(s.clientConfig.Region, q.Get("region")) // Region is actually not used as "en-US" is hardcoded if clientConfig.Region is empty
	// To be more precise, if s.clientConfig.Region is "US", it should be "US".
	// The original code defaults to "en-US" if c.region is "", but uses c.region if it's set.
	// Our test config has Region: "US", so it should be "US".

	// Check headers added by setupRequest
	s.Equal("application/json", capturedRequest.Header.Get("Content-Type"))
	s.Equal("application/json", capturedRequest.Header.Get("Accept"))
}

func (s *TMDBClientTestSuite) TestDoWithResponse_ClientError() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)

	client.SetHttpClient(s.mockHTTPClient) // Use the suite's mock client

	expectedError := errors.New("network error")
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, expectedError).Once()

	reqURL := client.GetBaseUrl().String() + "/test/path"
	req, err := http.NewRequest("GET", reqURL, nil)
	s.Require().NoError(err)

	// Act
	resp, err := client.DoWithResponse(req)

	// Assert
	s.Error(err)
	s.Nil(resp)
	s.EqualError(err, expectedError.Error())
}

func (s *TMDBClientTestSuite) TestDoWithRatingResponse_Success() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)
	client.SetHttpClient(s.mockHTTPClient)

	mockTMDBResp := tmdbModel.Response{
		Page: 1,
		Results: []tmdbModel.Entry{{
			ID:    123,
			Title: "Test Movie",
			Vote:  8.5,
		}},
	}
	mockBodyBytes, _ := json.Marshal(mockTMDBResp)
	mockRespBody := io.NopCloser(strings.NewReader(string(mockBodyBytes)))
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       mockRespBody,
		Header:     make(http.Header),
	}
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

	req, _ := http.NewRequest("GET", client.GetBaseUrl().String()+"/search/movie", nil)

	// Act
	ratingResp, err := client.DoWithRatingResponse(req)

	// Assert
	s.NoError(err)
	s.NotNil(ratingResp)
	parsedResp, ok := ratingResp.(*tmdbModel.Response)
	s.True(ok, "Response should be of type *tmdbModel.Response")
	s.Equal(mockTMDBResp.Page, parsedResp.Page)
	s.Len(parsedResp.Results, 1)
	s.Equal(mockTMDBResp.Results[0].ID, parsedResp.Results[0].ID)
	s.Equal(mockTMDBResp.Results[0].Title, parsedResp.Results[0].Title)
	s.Equal(mockTMDBResp.Results[0].Vote, parsedResp.Results[0].Vote)
}

func (s *TMDBClientTestSuite) TestDoWithRatingResponse_Unauthorized() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)
	client.SetHttpClient(s.mockHTTPClient)

	mockRespBody := io.NopCloser(strings.NewReader(`{"status_code":7,"status_message":"Invalid API key: You must be granted a valid key."}`))
	mockResponse := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       mockRespBody,
		Header:     make(http.Header),
	}
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

	req, _ := http.NewRequest("GET", client.GetBaseUrl().String()+"/search/movie", nil)

	// Act
	ratingResp, err := client.DoWithRatingResponse(req)

	// Assert
	s.Error(err)
	s.Nil(ratingResp)
	s.EqualError(err, model.NotAuthorized)
}

func (s *TMDBClientTestSuite) TestDoWithRatingResponse_NotFound() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)
	client.SetHttpClient(s.mockHTTPClient)

	mockRespBody := io.NopCloser(strings.NewReader(`{"status_code":34,"status_message":"The resource you requested could not be found."}`))
	mockResponse := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       mockRespBody,
		Header:     make(http.Header),
	}
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

	req, _ := http.NewRequest("GET", client.GetBaseUrl().String()+"/search/movie", nil)

	// Act
	ratingResp, err := client.DoWithRatingResponse(req)

	// Assert
	s.Error(err)
	s.Nil(ratingResp)
	s.EqualError(err, model.NotFound)
}

func (s *TMDBClientTestSuite) TestDoWithRatingResponse_InvalidJSON() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)
	client.SetHttpClient(s.mockHTTPClient)

	mockRespBody := io.NopCloser(strings.NewReader(`this is not json`))
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       mockRespBody,
		Header:     make(http.Header),
	}
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

	req, _ := http.NewRequest("GET", client.GetBaseUrl().String()+"/search/movie", nil)

	// Act
	ratingResp, err := client.DoWithRatingResponse(req)

	// Assert
	s.Error(err)
	s.Nil(ratingResp)
	// We don't assert the exact error message from json.Decode as it can be verbose
	s.Contains(err.Error(), "invalid character 'h' in literal true")
}

func (s *TMDBClientTestSuite) TestDoWithRatingResponse_DoRequestError() {
	// Arrange
	client, err := tmdbClient.NewTMDBClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)
	client.SetHttpClient(s.mockHTTPClient)

	expectedError := errors.New("underlying network error")
	s.mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, expectedError).Once()

	req, _ := http.NewRequest("GET", client.GetBaseUrl().String()+"/search/movie", nil)

	// Act
	ratingResp, err := client.DoWithRatingResponse(req)

	// Assert
	s.Error(err)
	s.Nil(ratingResp)
	s.EqualError(err, expectedError.Error())
}
