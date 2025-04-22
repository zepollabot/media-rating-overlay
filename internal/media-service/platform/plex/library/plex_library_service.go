package plex

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/samber/lo"
	"go.uber.org/zap"

	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	plex "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/model"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

// newRequestWithContextFunc defines the signature for a function that creates an HTTP request.
type newRequestWithContextFunc func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error)

type LibraryService interface {
	GetLibraries(ctx context.Context) ([]model.Library, error)
	RefreshLibrary(ctx context.Context, libraryID string, force bool) error
}

// PlexLibraryService handles Plex library operations
type PlexLibraryService struct {
	client         media.MediaClient
	logger         *zap.Logger
	NewRequestFunc newRequestWithContextFunc // Exported field for request creation
}

// NewPlexLibraryService creates a new Plex library service
func NewPlexLibraryService(client media.MediaClient, logger *zap.Logger) LibraryService {
	return &PlexLibraryService{
		client:         client,
		logger:         logger,
		NewRequestFunc: http.NewRequestWithContext, // Default to the real function
	}
}

// GetLibraries retrieves all libraries from Plex
func (s *PlexLibraryService) GetLibraries(ctx context.Context) ([]model.Library, error) {
	baseUrl := s.client.GetBaseUrl()
	endpoint := baseUrl.JoinPath("/library/sections")

	req, err := s.NewRequestFunc(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		s.logger.Error("unable to build request",
			zap.String("method", "GetLibraries"),
			zap.String("url", endpoint.String()),
			zap.Error(err),
		)
		return nil, err
	}

	response, err := s.client.DoWithMediaResponse(req)
	if err != nil {
		s.logger.Error("unable to perform request to Plex Client",
			zap.String("method", "GetLibraries"),
			zap.Error(err),
		)
		return nil, err
	}

	plexResponse, ok := response.(*plex.Response)
	if !ok {
		s.logger.Error("unable to cast response to plex.Response",
			zap.String("method", "GetLibraries"),
		)
		return nil, fmt.Errorf("invalid response type")
	}

	if len(plexResponse.MediaContainer.Libraries) == 0 {
		return []model.Library{}, nil
	}

	return s.convertPlexLibraries(plexResponse.MediaContainer.Libraries), nil
}

func (s *PlexLibraryService) RefreshLibrary(ctx context.Context, libraryID string, force bool) error {
	baseUrl := s.client.GetBaseUrl()
	endpoint := baseUrl.JoinPath(fmt.Sprintf("/library/sections/%s/refresh", libraryID))

	req, err := s.NewRequestFunc(ctx, http.MethodGet, endpoint.String(), nil)

	if err != nil {
		s.logger.Error("unable to build request",
			zap.String("method", "RefreshLibrary"),
			zap.String("url", endpoint.String()),
			zap.Error(err),
		)
		return err
	}

	if force {
		q := req.URL.Query()
		q.Add("force", strconv.Itoa(1))
		req.URL.RawQuery = q.Encode()
	}

	_, err = s.client.DoWithResponse(req)

	if err != nil {
		s.logger.Error("unable to perform request to Plex Client",
			zap.String("method", "RefreshLibrary"),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// convertPlexLibraries converts Plex library response to common Library model
func (s *PlexLibraryService) convertPlexLibraries(libraries []plex.Library) []model.Library {
	var convertedLibraries []model.Library
	lo.ForEach(libraries, func(item plex.Library, index int) {
		convertedLibraries = append(convertedLibraries, model.Library{
			ID:       item.ID,
			Type:     item.Type,
			Name:     item.Title,
			Language: item.Language,
		})
	})
	return convertedLibraries
}
