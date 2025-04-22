package plex_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	configmodel "github.com/zepollabot/media-rating-overlay/internal/config/model"
	mediamock "github.com/zepollabot/media-rating-overlay/internal/media-service/mocks"
	plexposter "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/poster"
	"github.com/zepollabot/media-rating-overlay/internal/model"
	storagemock "github.com/zepollabot/media-rating-overlay/internal/ports/mocks"
)

type PlexPosterServiceTestSuite struct {
	suite.Suite
	mockClient  *mediamock.MediaClient
	mockStorage *storagemock.PosterStorage
	service     *plexposter.PlexPosterService // Store as pointer to concrete type
	logger      *zap.Logger
	baseURL     *url.URL
}

func (s *PlexPosterServiceTestSuite) SetupTest() {
	s.mockClient = mediamock.NewMediaClient(s.T())
	s.mockStorage = storagemock.NewPosterStorage(s.T())
	s.logger = zap.NewNop()
	s.baseURL, _ = url.Parse("http://plex.test:32400")

	// NewPlexPosterService returns media.PosterService, so we cast it to the concrete type
	// to access NewRequestFunc for overriding in tests.
	service, ok := plexposter.NewPlexPosterService(s.mockClient, s.logger, s.mockStorage).(*plexposter.PlexPosterService)
	assert.True(s.T(), ok, "Failed to cast service to *plexposter.PlexPosterService")
	s.service = service

	s.mockClient.On("GetBaseUrl").Maybe().Return(s.baseURL)
}

func (s *PlexPosterServiceTestSuite) TearDownTest() {
	// Restore the original NewRequestFunc in case it was changed by a test
	s.service.NewRequestFunc = http.NewRequestWithContext
	s.mockClient.AssertExpectations(s.T())
	s.mockStorage.AssertExpectations(s.T())
}

func TestPlexPosterServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PlexPosterServiceTestSuite))
}

func (s *PlexPosterServiceTestSuite) TestGetPosterDiskPosition_ExistingPosterFound() {
	// Arrange
	ctx := context.Background()
	item := model.Item{
		ID: "1",
		Media: []model.Media{
			{File: []model.File{{Position: "/mnt/movies/Movie Title (2023)/Movie Title (2023).mkv"}}},
		},
	}
	libConfig := &configmodel.Library{Path: "/mnt/movies"}
	expectedPosterPath := "/mnt/movies/Movie Title (2023)/Movie Title (2023)-original.jpeg"

	s.mockStorage.On("CheckIfPosterExists", expectedPosterPath).Return(true, nil).Once()

	// Act
	pos, err := s.service.GetPosterDiskPosition(ctx, item, libConfig)

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedPosterPath, pos)
	s.mockStorage.AssertCalled(s.T(), "CheckIfPosterExists", expectedPosterPath)
}

func (s *PlexPosterServiceTestSuite) TestGetPosterDiskPosition_NoExistingPoster_DownloadsAndCalculates() {
	// Arrange
	ctx := context.Background()
	posterURL := "/library/metadata/1/thumb/12345.jpg"
	item := model.Item{
		ID:     "1",
		Poster: posterURL,
		Media:  []model.Media{{File: []model.File{{Position: "/mnt/tv/Show Name/Season 01/Show Name S01E01.mkv"}}}},
	}
	libConfig := &configmodel.Library{Path: "/mnt/tv"}

	// Mock finding no existing jpeg or png
	expectedJpegPath := "/mnt/tv/Show Name/Season 01/Show Name S01E01-original.jpeg"
	expectedPngPath := "/mnt/tv/Show Name/Season 01/Show Name S01E01-original.png"
	s.mockStorage.On("CheckIfPosterExists", expectedJpegPath).Return(false, nil).Once()
	s.mockStorage.On("CheckIfPosterExists", expectedPngPath).Return(false, nil).Once()

	// Mock poster download
	posterData := []byte("jpeg_image_data_here") // Actual JPEG data would be needed for correct mime type detection
	// Forcing a specific mime type for test stability rather than relying on DetectContentType with dummy data.
	// In a real scenario, provide actual image data.
	// http.DetectContentType is sensitive to the first 512 bytes.
	// To simulate JPEG: header should be something like FF D8 FF E0
	jfifHeader := []byte{0xff, 0xd8, 0xff, 0xe0}
	posterData = append(jfifHeader, posterData...)

	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(posterData))),
	}
	s.mockClient.On("DoWithResponse", mock.AnythingOfType("*http.Request")).Return(mockResp, nil).Once()

	expectedFinalPath := expectedJpegPath // Since we are returning jpeg data

	// Act
	pos, err := s.service.GetPosterDiskPosition(ctx, item, libConfig)

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedFinalPath, pos)
}

func (s *PlexPosterServiceTestSuite) TestGetPosterDiskPosition_StorageCheckError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Media: []model.Media{{File: []model.File{{Position: "/movies/A/a.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}
	expectedErr := errors.New("storage check failed")

	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, expectedErr).Once()

	// Act
	pos, err := s.service.GetPosterDiskPosition(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
	assert.Empty(s.T(), pos)
}

func (s *PlexPosterServiceTestSuite) TestGetPosterDiskPosition_GetPosterRequestCreationError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Poster: "/some/poster.jpg", Media: []model.Media{{File: []model.File{{Position: "/movies/A/a.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}
	expectedErr := errors.New("request creation failed")

	// Mock finding no existing posters
	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, nil).Twice() // .jpeg and .png

	// Mock NewRequestFunc to return an error
	originalNewRequestFunc := s.service.NewRequestFunc
	s.service.NewRequestFunc = func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
		return nil, expectedErr
	}
	defer func() { s.service.NewRequestFunc = originalNewRequestFunc }()

	// Act
	pos, err := s.service.GetPosterDiskPosition(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
	assert.Empty(s.T(), pos)
}

func (s *PlexPosterServiceTestSuite) TestGetPosterDiskPosition_GetPosterClientDoError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Poster: "/some/poster.jpg", Media: []model.Media{{File: []model.File{{Position: "/movies/A/a.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}
	expectedErr := errors.New("client.Do failed")

	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, nil).Twice()
	s.mockClient.On("DoWithResponse", mock.AnythingOfType("*http.Request")).Return(nil, expectedErr).Once()

	// Act
	pos, err := s.service.GetPosterDiskPosition(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
	assert.Empty(s.T(), pos)
}

func (s *PlexPosterServiceTestSuite) TestGetPosterDiskPosition_UnsupportedMimeType() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Poster: "/some/poster.webp", Media: []model.Media{{File: []model.File{{Position: "/movies/A/a.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}

	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, nil).Twice()

	// Simulate poster data for an unsupported type (e.g., text file)
	posterData := []byte("this is not an image")
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(posterData))),
	}
	s.mockClient.On("DoWithResponse", mock.AnythingOfType("*http.Request")).Return(mockResp, nil).Once()

	// Act
	pos, err := s.service.GetPosterDiskPosition(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "unable to find extension for mime type text/plain")
	assert.Empty(s.T(), pos)
}

func (s *PlexPosterServiceTestSuite) TestGetPosterDiskPosition_PathNotInConfig() {
	// Arrange
	ctx := context.Background()
	item := model.Item{
		Poster: "/library/metadata/1/thumb/12345.jpg",
		Media:  []model.Media{{File: []model.File{{Position: "/other/movies/Movie Title/Movie.mkv"}}}},
	}
	// Library config path is /mnt/plex_media, but media file is in /other/movies
	libConfig := &configmodel.Library{Path: "/mnt/plex_media"}

	// Act
	pos, err := s.service.GetPosterDiskPosition(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	// The error comes from s.getFilePosition when strings.Cut fails to find config.Path in fileDir
	assert.EqualError(s.T(), err, "unable to find file")
	assert.Empty(s.T(), pos)
	s.mockStorage.AssertNotCalled(s.T(), "CheckIfPosterExists", mock.AnythingOfType("string"))
	s.mockClient.AssertNotCalled(s.T(), "DoWithResponse", mock.AnythingOfType("*http.Request"))
	s.mockStorage.AssertNotCalled(s.T(), "SavePoster", mock.Anything, mock.Anything)
}

func (s *PlexPosterServiceTestSuite) TestEnsurePosterExists_PosterAlreadyExists() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Media: []model.Media{{File: []model.File{{Position: "/movies/A/a.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}
	expectedPath := "/movies/A/a-original.jpeg"

	s.mockStorage.On("CheckIfPosterExists", expectedPath).Return(true, nil).Once()

	// Act
	err := s.service.EnsurePosterExists(ctx, item, libConfig)

	// Assert
	assert.NoError(s.T(), err)
	s.mockStorage.AssertCalled(s.T(), "CheckIfPosterExists", expectedPath)
	// Ensure SavePoster is not called
	s.mockStorage.AssertNotCalled(s.T(), "SavePoster", mock.Anything, mock.Anything)
}

func (s *PlexPosterServiceTestSuite) TestEnsurePosterExists_DownloadsAndSavesNewPoster() {
	// Arrange
	ctx := context.Background()
	posterURL := "/library/metadata/2/thumb/67890.png"
	item := model.Item{
		ID:     "2",
		Poster: posterURL,
		Media:  []model.Media{{File: []model.File{{Position: "/series/B/S01/b_s01e01.mp4"}}}},
	}
	libConfig := &configmodel.Library{Path: "/series"}

	// Mock finding no existing posters (.jpeg, .png)
	expectedJpegPath := "/series/B/S01/b_s01e01-original.jpeg"
	expectedPngPath := "/series/B/S01/b_s01e01-original.png"
	s.mockStorage.On("CheckIfPosterExists", expectedJpegPath).Return(false, nil).Once()
	s.mockStorage.On("CheckIfPosterExists", expectedPngPath).Return(false, nil).Once()

	// Mock poster download (simulating PNG data)
	pngHeader := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a} // Minimal PNG header
	posterData := append(pngHeader, []byte("actual_png_image_data")...)
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(posterData))),
	}
	s.mockClient.On("DoWithResponse", mock.AnythingOfType("*http.Request")).Return(mockResp, nil).Once()

	// Mock saving the poster
	// getFilePosition will calculate this based on the detected .png extension
	expectedSavePath := expectedPngPath
	s.mockStorage.On("SavePoster", expectedSavePath, posterData).Return(nil).Once()

	// Act
	err := s.service.EnsurePosterExists(ctx, item, libConfig)

	// Assert
	assert.NoError(s.T(), err)
	s.mockStorage.AssertCalled(s.T(), "SavePoster", expectedSavePath, posterData)
}

func (s *PlexPosterServiceTestSuite) TestEnsurePosterExists_StorageCheckError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Media: []model.Media{{File: []model.File{{Position: "/movies/C/c.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}
	expectedErr := errors.New("storage check error during ensure")

	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, expectedErr).Once()

	// Act
	err := s.service.EnsurePosterExists(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
}

func (s *PlexPosterServiceTestSuite) TestEnsurePosterExists_GetPosterFails_RequestCreationError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Poster: "/fail/poster.jpg", Media: []model.Media{{File: []model.File{{Position: "/movies/D/d.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}
	expectedErr := errors.New("ensure: request creation failed")

	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, nil).Twice()

	originalNewRequestFunc := s.service.NewRequestFunc
	s.service.NewRequestFunc = func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
		return nil, expectedErr
	}
	defer func() { s.service.NewRequestFunc = originalNewRequestFunc }()

	// Act
	err := s.service.EnsurePosterExists(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
}

func (s *PlexPosterServiceTestSuite) TestEnsurePosterExists_GetPosterFails_ClientDoError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Poster: "/fail/poster.jpg", Media: []model.Media{{File: []model.File{{Position: "/movies/E/e.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}
	expectedErr := errors.New("ensure: client.Do failed")

	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, nil).Twice()
	s.mockClient.On("DoWithResponse", mock.AnythingOfType("*http.Request")).Return(nil, expectedErr).Once()

	// Act
	err := s.service.EnsurePosterExists(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
}

func (s *PlexPosterServiceTestSuite) TestEnsurePosterExists_UnsupportedMimeTypePreventsSave() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Poster: "/poster.txt", Media: []model.Media{{File: []model.File{{Position: "/movies/F/f.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}

	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, nil).Twice()

	posterData := []byte("this is plain text, not an image")
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(posterData))),
	}
	s.mockClient.On("DoWithResponse", mock.AnythingOfType("*http.Request")).Return(mockResp, nil).Once()

	// Act
	err := s.service.EnsurePosterExists(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "unable to find extension for mime type text/plain")
	s.mockStorage.AssertNotCalled(s.T(), "SavePoster", mock.Anything, mock.Anything)
}

func (s *PlexPosterServiceTestSuite) TestEnsurePosterExists_SavePosterError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{Poster: "/poster.jpg", Media: []model.Media{{File: []model.File{{Position: "/movies/G/g.mkv"}}}}}
	libConfig := &configmodel.Library{Path: "/movies"}
	expectedErr := errors.New("failed to save poster")

	s.mockStorage.On("CheckIfPosterExists", mock.AnythingOfType("string")).Return(false, nil).Twice()

	jfifHeader := []byte{0xff, 0xd8, 0xff, 0xe0}
	posterData := append(jfifHeader, []byte("dummy jpeg data")...)
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(posterData))),
	}
	s.mockClient.On("DoWithResponse", mock.AnythingOfType("*http.Request")).Return(mockResp, nil).Once()

	expectedSavePath := "/movies/G/g-original.jpeg"
	s.mockStorage.On("SavePoster", expectedSavePath, posterData).Return(expectedErr).Once()

	// Act
	err := s.service.EnsurePosterExists(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
}

func (s *PlexPosterServiceTestSuite) TestEnsurePosterExists_PathNotInConfig() {
	// Arrange
	ctx := context.Background()
	item := model.Item{
		Poster: "/library/metadata/1/thumb/12345.jpg",
		Media:  []model.Media{{File: []model.File{{Position: "/other/movies/Movie Title/Movie.mkv"}}}},
	}
	libConfig := &configmodel.Library{Path: "/mnt/plex_media"}
	// Act
	err := s.service.EnsurePosterExists(ctx, item, libConfig)

	// Assert
	assert.Error(s.T(), err)
	assert.EqualError(s.T(), err, "unable to find file")
	s.mockStorage.AssertNotCalled(s.T(), "CheckIfPosterExists", mock.AnythingOfType("string"))
	s.mockStorage.AssertNotCalled(s.T(), "SavePoster", mock.Anything, mock.Anything)
	s.mockClient.AssertNotCalled(s.T(), "DoWithResponse", mock.AnythingOfType("*http.Request"))
}
