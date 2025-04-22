package tmdb

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	common "github.com/zepollabot/media-rating-overlay/internal/httpclient"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
	tmdb "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/model"
)

// TMDBClient implements the RatingClient interface
type TMDBClient struct {
	httpClient common.ServiceHTTPClient
	apiKey     string
	language   string
	region     string
	baseUrl    url.URL
	logger     *zap.Logger
}

// NewTMDBClient creates a new TMDB client
func NewTMDBClient(clientConfig *config.TMDB, httpClientConfig *config.HTTPClient, logFilePath string, logger *zap.Logger) (*TMDBClient, error) {
	if clientConfig.ApiKey == "" {
		logger.Error("tmdb.api_key is required")
		return nil, errors.New("tmdb.api_key is required")
	}

	baseUrl := url.URL{
		Scheme: "https",
		Host:   "api.themoviedb.org",
	}

	httpClient := NewTMDBHTTPClient(httpClientConfig.Timeout, httpClientConfig.MaxRetries)

	if err := common.SetupLogging(httpClient.client, logFilePath); err != nil {
		return nil, err
	}

	return &TMDBClient{
		httpClient: httpClient,
		apiKey:     clientConfig.ApiKey,
		language:   clientConfig.Language,
		region:     clientConfig.Region,
		baseUrl:    baseUrl,
		logger:     logger,
	}, nil
}

func (c *TMDBClient) DoWithResponse(request *http.Request) (*http.Response, error) {
	if err := c.setupRequest(request); err != nil {
		c.logger.Error("unable to setup request", zap.Error(err))
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error("unable to perform request to TMDB API",
			zap.String("url", request.URL.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return resp, nil
}

// DoWithRatingResponse performs a request and returns a parsed TMDB response
func (c *TMDBClient) DoWithRatingResponse(request *http.Request) (rating.RatingResponse, error) {
	resp, err := c.DoWithResponse(request)
	if err != nil {
		return nil, err
	}

	return c.parseTMDBResponse(resp)
}

// setupRequest configures the request with common headers and authentication
func (c *TMDBClient) setupRequest(request *http.Request) error {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	// Add TMDB api key to every request
	q := request.URL.Query()
	q.Add("api_key", c.apiKey)

	language := "en"
	if c.language != "" {
		language = c.language
	}
	q.Add("language", language)

	region := "en-US"
	if c.region != "" {
		region = c.region
	}
	q.Add("region", region)

	request.URL.RawQuery = q.Encode()

	return nil
}

// parseTMDBResponse handles the TMDB API response parsing
func (c *TMDBClient) parseTMDBResponse(resp *http.Response) (*tmdb.Response, error) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("error closing response body", zap.Error(err))
		}
	}()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		err := errors.New(model.NotAuthorized)
		c.logger.Error(
			"your TMDB Api key is invalid or expired, please use a valid Api key",
			zap.Error(err),
		)
		return nil, err
	case http.StatusNotFound:
		err := errors.New(model.NotFound)
		c.logger.Error(
			"cannot find the resource, please check the query",
			zap.Error(err),
		)
		return nil, err
	}

	var TMDBResponse tmdb.Response
	err := json.NewDecoder(resp.Body).Decode(&TMDBResponse)
	if err != nil {
		c.logger.Error("unable to decode TMDB API response",
			zap.Error(err),
		)
		return nil, err
	}

	return &TMDBResponse, nil
}

func (c *TMDBClient) GetBaseUrl() *url.URL {
	return &c.baseUrl
}

// SetHttpClient replaces the internal HttpClient with the provided one
// Method used primarily for testing
func (c *TMDBClient) SetHttpClient(client common.ServiceHTTPClient) {
	c.httpClient = client
}
