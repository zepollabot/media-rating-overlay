package metacritic_test

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
	metacriticClient "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/metacritic/client"
	metacriticModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/metacritic/model"
)

type MetacriticClientTestSuite struct {
	suite.Suite
	mockHTTPClient   *httpClientMocks.ServiceHTTPClient
	logger           *zap.Logger
	clientConfig     *config.Metacritic
	httpClientConfig *config.HTTPClient
}

func (s *MetacriticClientTestSuite) SetupTest() {
	s.mockHTTPClient = httpClientMocks.NewServiceHTTPClient(s.T())
	s.logger = zap.NewNop() // Use zap.NewExample() or zap.NewDevelopment() for verbose logs
	s.clientConfig = &config.Metacritic{
		APIKey: "test-api-key",
	}
	s.httpClientConfig = &config.HTTPClient{
		Timeout:    5, // 5 seconds
		MaxRetries: 3,
	}
}

func (s *MetacriticClientTestSuite) TearDownTest() {
	s.mockHTTPClient.AssertExpectations(s.T())
}

func TestMetacriticClientTestSuite(t *testing.T) {
	suite.Run(t, new(MetacriticClientTestSuite))
}

// TestNewTMDBClient_Success tests the successful creation of a new TMDBClient
func (s *MetacriticClientTestSuite) TestNewTMDBClient_Success() {
	// Arrange
	// SetupTest already prepares clientConfig, httpClientConfig, and logger

	// Act
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)

	// Assert
	s.NoError(err)
	s.NotNil(client)
	// Basic check, more detailed checks can be done on methods using these fields
	s.Equal("https", client.GetBaseUrl().Scheme)
	s.Equal("backend.metacritic.com", client.GetBaseUrl().Host)
}

func (s *MetacriticClientTestSuite) TestGetBaseUrl() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)

	// Act
	baseUrl := client.GetBaseUrl()

	// Assert
	s.NotNil(baseUrl)
	s.Equal("https", baseUrl.Scheme)
	s.Equal("backend.metacritic.com", baseUrl.Host)
}

func (s *MetacriticClientTestSuite) TestSetHttpClient() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
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

func (s *MetacriticClientTestSuite) TestDoWithResponse_Success() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
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
	s.Equal(s.clientConfig.APIKey, q.Get("apiKey"))
	s.Equal("2", q.Get("mcoTypeId"))
	s.Equal("DESC", q.Get("sortDirection"))
	s.Equal("24", q.Get("limit"))
	s.Equal("0", q.Get("offset"))

	// Check headers added by setupRequest
	s.Equal("application/json", capturedRequest.Header.Get("Content-Type"))
	s.Equal("application/json", capturedRequest.Header.Get("Accept"))
}

func (s *MetacriticClientTestSuite) TestDoWithResponse_ClientError() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
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

func (s *MetacriticClientTestSuite) TestDoWithRatingResponse_Success() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
	s.Require().NoError(err)
	s.Require().NotNil(client)
	client.SetHttpClient(s.mockHTTPClient)

	mockTMDBResp := metacriticModel.Response{
		Data: metacriticModel.Data{
			TotalResults: 1,
			ID:           "123",
			Items: []metacriticModel.Entry{{
				ID:     123,
				Title:  "Test Movie",
				Rating: "85",
			}},
		},
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
	parsedResp, ok := ratingResp.(*metacriticModel.Response)
	s.True(ok, "Response should be of type *metacriticModel.Response")
	s.Equal(mockTMDBResp.Data.TotalResults, parsedResp.Data.TotalResults)
	s.Len(parsedResp.Data.Items, 1)
	s.Equal(mockTMDBResp.Data.Items[0].ID, parsedResp.Data.Items[0].ID)
	s.Equal(mockTMDBResp.Data.Items[0].Title, parsedResp.Data.Items[0].Title)
	s.Equal(mockTMDBResp.Data.Items[0].Rating, parsedResp.Data.Items[0].Rating)
}

func (s *MetacriticClientTestSuite) TestDoWithRatingResponse_Unauthorized() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
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

func (s *MetacriticClientTestSuite) TestDoWithRatingResponse_NotFound() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
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

func (s *MetacriticClientTestSuite) TestDoWithRatingResponse_InvalidJSON() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
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

func (s *MetacriticClientTestSuite) TestDoWithRatingResponse_DoRequestError() {
	// Arrange
	client, err := metacriticClient.NewMetacriticClient(s.clientConfig, s.httpClientConfig, os.DevNull, s.logger)
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
