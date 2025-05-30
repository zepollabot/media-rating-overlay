package metacritic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	common "github.com/zepollabot/media-rating-overlay/internal/httpclient"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
	metacritic "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/metacritic/model"
)

// MetacriticClient implements the RatingClient interface
type MetacriticClient struct {
	httpClient common.ServiceHTTPClient
	apiKey     string
	baseUrl    url.URL
	logger     *zap.Logger
}

// NewMetacriticClient creates a new Metacritic client
func NewMetacriticClient(clientConfig *config.Metacritic, httpClientConfig *config.HTTPClient, logFilePath string, logger *zap.Logger) (*MetacriticClient, error) {
	if clientConfig.APIKey == "" {
		logger.Error("metacritic.api_key is required")
		return nil, errors.New("metacritic.api_key is required")
	}

	baseUrl := url.URL{
		Scheme: "https",
		Host:   "backend.metacritic.com",
	}

	httpClient := NewMetacriticHTTPClient(httpClientConfig.Timeout, httpClientConfig.MaxRetries)

	if err := common.SetupLogging(httpClient.client, logFilePath); err != nil {
		return nil, err
	}

	return &MetacriticClient{
		httpClient: httpClient,
		apiKey:     clientConfig.APIKey,
		baseUrl:    baseUrl,
		logger:     logger,
	}, nil
}

func (c *MetacriticClient) DoWithResponse(request *http.Request) (*http.Response, error) {
	if err := c.setupRequest(request); err != nil {
		c.logger.Error("unable to setup request", zap.Error(err))
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error("unable to perform request to Metacritic",
			zap.String("url", request.URL.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return resp, nil
}

// DoWithRatingResponse performs a request and returns a parsed Metacritic response
func (c *MetacriticClient) DoWithRatingResponse(request *http.Request) (rating.RatingResponse, error) {
	resp, err := c.DoWithResponse(request)
	if err != nil {
		return nil, err
	}

	return c.parseMetacriticResponse(resp)
}

// setupRequest configures the request with common headers
func (c *MetacriticClient) setupRequest(request *http.Request) error {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	// Add TMDB api key to every request
	q := request.URL.Query()
	q.Add("apiKey", c.apiKey)

	// Add product type to every request
	q.Add("mcoTypeId", strconv.Itoa(metacritic.MetacriticProductTypeMovie))

	// Add sort by relevance to every request
	q.Add("sortDirection", "DESC")

	// Add limit to every request
	q.Add("limit", "24")
	q.Add("offset", "0")

	request.URL.RawQuery = q.Encode()

	return nil
}

// parseMetacriticResponse handles the Metacritic response parsing
func (c *MetacriticClient) parseMetacriticResponse(resp *http.Response) (*metacritic.Response, error) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("error closing response body", zap.Error(err))
		}
	}()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		err := errors.New(model.NotAuthorized)
		c.logger.Error(
			"your Metacritic Api key is invalid or expired, please use a valid Api key",
			zap.Error(err),
		)
		return nil, err
	case http.StatusNotFound:
		err := errors.New(model.NotFound)
		c.logger.Debug(
			"cannot find the resource, please check the query",
			zap.Error(err),
		)
		return nil, err
	case http.StatusOK:
		var response metacritic.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			c.logger.Error("failed to decode response body",
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &response, nil
	default:
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		c.logger.Error("unexpected status code",
			zap.Int("status_code", resp.StatusCode),
			zap.Error(err),
		)
		return nil, err
	}
}

func (c *MetacriticClient) GetBaseUrl() *url.URL {
	return &c.baseUrl
}

// SetHttpClient replaces the internal HttpClient with the provided one
// Method used primarily for testing
func (c *MetacriticClient) SetHttpClient(client common.ServiceHTTPClient) {
	c.httpClient = client
}
