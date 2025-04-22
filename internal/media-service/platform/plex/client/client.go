package plex

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	common "github.com/zepollabot/media-rating-overlay/internal/httpclient"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	plexModels "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/model"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
)

// PlexClient implements the MediaClient interface
type PlexClient struct {
	httpClient common.ServiceHTTPClient
	token      string
	baseUrl    url.URL
	logger     *zap.Logger
}

// NewPlexClient creates a new Plex client
func NewPlexClient(clientConfig *config.Plex, httpClientConfig *config.HTTPClient, logFilePath string, logger *zap.Logger) (*PlexClient, error) {
	if clientConfig.Url == "" {
		logger.Error("plex.url is required")
		return nil, errors.New("plex.url is required")
	}

	baseUrl, err := url.Parse(clientConfig.Url)
	if err != nil {
		logger.Error("error parsing Plex URL", zap.Error(err))
		return nil, err
	}

	httpClient := NewPlexHTTPClient(httpClientConfig.Timeout, httpClientConfig.MaxRetries)

	if err := common.SetupLogging(httpClient.client, logFilePath); err != nil {
		logger.Error("error setting up logging", zap.Error(err))
		return nil, err
	}

	return &PlexClient{
		httpClient: httpClient,
		token:      clientConfig.Token,
		baseUrl:    *baseUrl,
		logger:     logger,
	}, nil
}

// DoWithResponse performs a request and returns the raw HTTP response
func (c *PlexClient) DoWithResponse(request *http.Request) (*http.Response, error) {
	if err := c.setupRequest(request); err != nil {
		c.logger.Error("unable to setup request", zap.Error(err))
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Error("unable to perform request to Plex API",
			zap.String("url", request.URL.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return resp, nil
}

// DoWithMediaResponse performs a request and returns a parsed Plex response
func (c *PlexClient) DoWithMediaResponse(request *http.Request) (media.MediaResponse, error) {
	resp, err := c.DoWithResponse(request)
	if err != nil {
		return nil, err
	}

	return c.parsePlexResponse(resp)
}

// setupRequest configures the request with common headers and authentication
func (c *PlexClient) setupRequest(request *http.Request) error {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	q := request.URL.Query()
	q.Set("X-Plex-Token", c.token)
	request.URL.RawQuery = q.Encode()

	return nil
}

// parsePlexResponse handles the Plex API response parsing
func (c *PlexClient) parsePlexResponse(resp *http.Response) (*plexModels.Response, error) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("error closing response body", zap.Error(err))
		}
	}()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		err := errors.New(model.NotAuthorized)
		c.logger.Error(
			"your Plex token is invalid or expired, please use a valid Plex token",
			zap.Error(err),
		)
		return nil, err
	}

	var plexResponse plexModels.Response
	if err := json.NewDecoder(resp.Body).Decode(&plexResponse); err != nil {
		c.logger.Error("unable to decode Plex API response",
			zap.Error(err),
		)
		return nil, err
	}

	return &plexResponse, nil
}

// GetBaseUrl returns the base URL for the Plex server
func (c *PlexClient) GetBaseUrl() *url.URL {
	return &c.baseUrl
}

// SetHttpClient replaces the internal HttpClient (primarily for testing)
func (c *PlexClient) SetHttpClient(client common.HTTPClient) {
	c.httpClient = client
}
