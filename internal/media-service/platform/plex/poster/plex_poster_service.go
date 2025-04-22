package plex

import (
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	"github.com/zepollabot/media-rating-overlay/internal/model"
	"github.com/zepollabot/media-rating-overlay/internal/ports"
)

// newRequestWithContextFunc defines the signature for a function that creates an HTTP request.
type newRequestWithContextFunc func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error)

// PlexPosterService handles Plex poster operations
type PlexPosterService struct {
	client         media.MediaClient
	logger         *zap.Logger
	storage        ports.PosterStorage
	NewRequestFunc newRequestWithContextFunc // Exported field for request creation
}

// NewPlexPosterService creates a new Plex poster service
func NewPlexPosterService(client media.MediaClient, logger *zap.Logger, storage ports.PosterStorage) media.PosterService {
	return &PlexPosterService{
		client:         client,
		logger:         logger,
		storage:        storage,
		NewRequestFunc: http.NewRequestWithContext, // Default to the real function
	}
}

func (s *PlexPosterService) GetPosterDiskPosition(ctx context.Context, item model.Item, config *config.Library) (string, error) {
	s.logger.Debug("Getting poster disk position..",
		zap.String("Item ID", item.ID),
	)

	// First try to find an existing poster
	filePos, err := s.findExistingPoster(item.Media, config)
	if err != nil {
		return "", err
	}
	if filePos != "" {
		return filePos, nil
	}

	// If no existing poster found, determine the position for a new one
	posterData, err := s.getPoster(ctx, item.Poster)
	if err != nil {
		return "", err
	}

	// Detect the mime type and get extension
	posterMimeType := http.DetectContentType(posterData)
	posterFileExt, err := media.GetExtensionByMimeType(posterMimeType)
	if err != nil {
		return "", err
	}

	// Calculate the file position
	return s.getFilePosition(item.Media, posterFileExt, config)
}

func (s *PlexPosterService) EnsurePosterExists(ctx context.Context, item model.Item, config *config.Library) error {
	s.logger.Debug("Ensuring poster exists..",
		zap.String("Item ID", item.ID),
	)

	// First try to find an existing poster
	filePos, err := s.findExistingPoster(item.Media, config)
	if err != nil {
		return err
	}
	if filePos != "" {
		return nil // Poster already exists
	}

	// If no existing poster found, download and save a new one
	posterData, err := s.getPoster(ctx, item.Poster)
	if err != nil {
		return err
	}

	// Get the position for the new poster
	posterMimeType := http.DetectContentType(posterData)
	posterFileExt, err := media.GetExtensionByMimeType(posterMimeType)
	if err != nil {
		return err
	}

	filePos, err = s.getFilePosition(item.Media, posterFileExt, config)
	if err != nil {
		return err
	}

	return s.storage.SavePoster(filePos, posterData)
}

func (s *PlexPosterService) findExistingPoster(mediaModels []model.Media, config *config.Library) (string, error) {
	supportedExt := []string{".jpeg", ".png"}

	for _, ext := range supportedExt {
		posterFilePos, err := s.getFilePosition(mediaModels, ext, config)
		if err != nil {
			return "", err
		}

		found, err := s.storage.CheckIfPosterExists(posterFilePos)
		if err != nil {
			return "", err
		}

		if found {
			return posterFilePos, nil
		}
	}

	return "", nil
}

func (s *PlexPosterService) getPoster(ctx context.Context, posterURL string) ([]byte, error) {
	s.logger.Debug("Getting poster..",
		zap.String("posterURL", posterURL),
	)

	var poster []byte

	baseUrl := s.client.GetBaseUrl()
	endpoint := baseUrl.JoinPath(posterURL)

	req, err := s.NewRequestFunc(ctx, http.MethodGet, endpoint.String(), nil) // Use the injectable function

	if err != nil {
		s.logger.Error("unable to build request",
			zap.String("method", "getPoster"),
			zap.String("url", endpoint.String()),
			zap.Error(err),
		)
		return poster, err
	}

	resp, err := s.client.DoWithResponse(req)

	if err != nil {
		s.logger.Error("unable to perform request to Plex Client",
			zap.String("method", "getPoster"),
			zap.Error(err),
		)
		return poster, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.logger.Error("unable to close response body")
		}
	}(resp.Body)

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return poster, err
	}

	return bytes, nil
}

func (s *PlexPosterService) getFilePosition(mediaModels []model.Media, posterFileExt string, config *config.Library) (string, error) {
	s.logger.Debug("Getting file position..",
		zap.String("posterFileExt", posterFileExt),
		zap.String("library config path", config.Path),
	)

	var filePos string
	for _, mediaModel := range mediaModels {
		for _, file := range mediaModel.File {
			fileName := media.GetFileNameWithoutExtTrimSuffix(filepath.Base(file.Position))
			fileDir := filepath.Dir(file.Position)

			_, after, found := strings.Cut(fileDir, config.Path)

			if !found {
				s.logger.Error("unable to find config dir",
					zap.String("fileDir", fileDir),
					zap.String("library config path", config.Path),
				)
				return "", errors.New("unable to find file")
			}

			posterFileName := fileName + "-original" + posterFileExt

			if after != "" {
				filePos = config.Path + after + "/" + posterFileName
			} else {
				filePos = config.Path + "/" + posterFileName
			}
		}
	}
	return filePos, nil
}
