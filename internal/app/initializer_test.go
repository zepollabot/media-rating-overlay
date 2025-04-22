package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	mocks "github.com/zepollabot/media-rating-overlay/internal/app/mocks"
	configmodel "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/constant"
	factorymocks "github.com/zepollabot/media-rating-overlay/internal/factory/mocks"
	mediaModel "github.com/zepollabot/media-rating-overlay/internal/media-service/model"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

type ServiceInitializerSuite struct {
	suite.Suite
	logger        *zap.Logger
	mockConfig    *configmodel.Config
	ctx           context.Context
	workSemaphore *semaphore.Weighted

	initializer *ServiceInitializer
}

func (s *ServiceInitializerSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.mockConfig = &configmodel.Config{
		Plex: configmodel.Plex{
			Enabled: true,
			Url:     "http://dummy-plex-url",
			Token:   "dummy-plex-token",
		},
		TMDB: configmodel.TMDB{
			Enabled: true,
			ApiKey:  "dummy-tmdb-apikey",
		},
		Performance: configmodel.Performance{
			LibraryProcessingTimeout: 10,
		},
		HTTPClient: configmodel.HTTPClient{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		},
		Logger: configmodel.Logger{
			LogFilePath:    "test.log",
			LogLevel:       "info",
			MaxSize:        1,
			MaxBackups:     1,
			MaxAge:         1,
			Compress:       false,
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			UseJSON:        false,
			UseStdout:      true,
		},
		Processor: configmodel.ProcessorConfig{
			ItemProcessor: configmodel.ItemProcessorConfig{
				RatingBuilder: configmodel.RatingBuilderConfig{
					Timeout: 15 * time.Second,
				},
			},
			LibraryProcessor: configmodel.LibraryProcessorConfig{
				DefaultTimeout: 15 * time.Second,
			},
		},
	}
	s.ctx = context.Background()
	s.workSemaphore = semaphore.NewWeighted(10)

	s.initializer = NewServiceInitializer(s.logger, s.mockConfig, s.ctx, s.workSemaphore)
}

func (s *ServiceInitializerSuite) TearDownTest() {
}

func TestServiceInitializerSuite(t *testing.T) {
	suite.Run(t, new(ServiceInitializerSuite))
}

func (s *ServiceInitializerSuite) TestNewServiceInitializer() {
	assert.NotNil(s.T(), s.initializer)
	assert.Empty(s.T(), s.initializer.GetMediaServices())
	assert.Empty(s.T(), s.initializer.GetRatingPlatformServices())
	assert.Nil(s.T(), s.initializer.GetLibraryProcessor())
	assert.Nil(s.T(), s.initializer.GetItemProcessor())
}

func (s *ServiceInitializerSuite) TestInitializeServices_Success() {
	err := s.initializer.InitializeServices()
	assert.NoError(s.T(), err)

	assert.NotNil(s.T(), s.initializer.GetMediaServices())
	assert.Len(s.T(), s.initializer.GetMediaServices(), 1)
	assert.Equal(s.T(), mediaModel.MediaServicePlex, s.initializer.GetMediaServices()[0].Name)

	assert.NotNil(s.T(), s.initializer.GetRatingPlatformServices())
	assert.Len(s.T(), s.initializer.GetRatingPlatformServices(), 3)

	ratingServicesNames := make([]string, 0, len(s.initializer.GetRatingPlatformServices()))
	for _, rs := range s.initializer.GetRatingPlatformServices() {
		ratingServicesNames = append(ratingServicesNames, rs.Name)
	}
	assert.Contains(s.T(), ratingServicesNames, constant.RatingServiceTMDB)
	assert.Contains(s.T(), ratingServicesNames, constant.RatingServiceRottenTomatoes)
	assert.Contains(s.T(), ratingServicesNames, constant.RatingServiceIMDB)

	assert.NotNil(s.T(), s.initializer.GetLibraryProcessor())
	assert.NotNil(s.T(), s.initializer.GetItemProcessor())
}

func (s *ServiceInitializerSuite) TestGetMediaServices_BeforeInitialization() {
	services := s.initializer.GetMediaServices()
	assert.Empty(s.T(), services)
}

func (s *ServiceInitializerSuite) TestGetMediaServices_AfterInitialization() {
	err := s.initializer.InitializeServices()
	s.Require().NoError(err)
	services := s.initializer.GetMediaServices()
	assert.NotEmpty(s.T(), services)
	assert.Len(s.T(), services, 1)
	assert.Equal(s.T(), mediaModel.MediaServicePlex, services[0].Name)
}

func (s *ServiceInitializerSuite) TestGetRatingPlatformServices_BeforeInitialization() {
	services := s.initializer.GetRatingPlatformServices()
	assert.Empty(s.T(), services)
}

func (s *ServiceInitializerSuite) TestGetRatingPlatformServices_AfterInitialization() {
	err := s.initializer.InitializeServices()
	s.Require().NoError(err)
	services := s.initializer.GetRatingPlatformServices()
	assert.NotEmpty(s.T(), services)
	assert.Len(s.T(), services, 3)
}

func (s *ServiceInitializerSuite) TestGetLibraryProcessor_BeforeInitialization() {
	processor := s.initializer.GetLibraryProcessor()
	assert.Nil(s.T(), processor)
}

func (s *ServiceInitializerSuite) TestGetLibraryProcessor_AfterInitialization() {
	err := s.initializer.InitializeServices()
	s.Require().NoError(err)
	processor := s.initializer.GetLibraryProcessor()
	assert.NotNil(s.T(), processor)
}

func (s *ServiceInitializerSuite) TestGetItemProcessor_BeforeInitialization() {
	processor := s.initializer.GetItemProcessor()
	assert.Nil(s.T(), processor)
}

func (s *ServiceInitializerSuite) TestGetItemProcessor_AfterInitialization() {
	err := s.initializer.InitializeServices()
	s.Require().NoError(err)
	processor := s.initializer.GetItemProcessor()
	assert.NotNil(s.T(), processor)
}

func (s *ServiceInitializerSuite) TestInitializeServices_PlexDisabled() {
	s.mockConfig.Plex.Enabled = false // Disable Plex
	// Re-initialize with the modified config
	s.initializer = NewServiceInitializer(s.logger, s.mockConfig, s.ctx, s.workSemaphore)

	err := s.initializer.InitializeServices()
	assert.NoError(s.T(), err)

	assert.Empty(s.T(), s.initializer.GetMediaServices()) // No media services if Plex is disabled

	// Rating services and processors should still be initialized
	assert.NotNil(s.T(), s.initializer.GetRatingPlatformServices())
	assert.Len(s.T(), s.initializer.GetRatingPlatformServices(), 3)
	assert.NotNil(s.T(), s.initializer.GetLibraryProcessor())
	assert.NotNil(s.T(), s.initializer.GetItemProcessor())
}

func (s *ServiceInitializerSuite) TestInitializeServices_MediaServiceCreationError() {
	// Arrange
	mockMediaServiceModelFactory := new(mocks.MediaServiceModelFactory)

	// Use the s.initializer from SetupTest, but replace its factory with the mock.
	// This now works because MediaServiceModelFactoryExported is an exported field.
	s.initializer.MediaServiceModelFactory = mockMediaServiceModelFactory

	expectedError := errors.New("failed to create media service")
	// Ensure Plex is enabled in the config for this test path to be hit.
	s.mockConfig.Plex.Enabled = true
	mockMediaServiceModelFactory.On("Create", mediaModel.MediaServicePlex).Return(mediaModel.MediaService{}, expectedError).Once()

	// Act
	err := s.initializer.InitializeServices()

	// Assert
	assert.Error(s.T(), err)
	assert.EqualError(s.T(), err, expectedError.Error())
	mockMediaServiceModelFactory.AssertExpectations(s.T())
	assert.Empty(s.T(), s.initializer.GetMediaServices(), "Media services should be empty on creation error")
}

func (s *ServiceInitializerSuite) TestInitializeServices_RatingServiceCreationError() {
	// Arrange
	mockRatingServiceBaseFactory := new(factorymocks.RatingServiceBaseFactory)
	s.initializer.RatingServiceBaseFactory = mockRatingServiceBaseFactory

	expectedError := errors.New("failed to create rating service component")

	mockRatingServiceBaseFactory.On("BuildTMDBComponents").Return(ratingModel.RatingService{}, expectedError).Once()

	// Act
	err := s.initializer.InitializeServices()

	// Assert
	assert.Error(s.T(), err)
	// The error from BuildTMDBComponents is wrapped by ratingPlatformServiceModelFactory.Create
	// and then by buildRatingPlatformServicesArray.
	assert.Contains(s.T(), err.Error(), expectedError.Error(), "Error message should contain the original error")
	mockRatingServiceBaseFactory.AssertExpectations(s.T())
	assert.Empty(s.T(), s.initializer.GetRatingPlatformServices(), "Rating platform services should be empty on creation error")
}

// Add config-based error tests as previously discussed, they are still valid and good to have.
func (s *ServiceInitializerSuite) TestInitializeServices_MediaServiceConfigError_PlexUrlMissing() {
	// Arrange
	invalidConfig := &configmodel.Config{
		Plex: configmodel.Plex{
			Enabled: true,
			Url:     "", // Plex URL is missing
			Token:   "dummy-token",
		},
		TMDB:        configmodel.TMDB{Enabled: true, ApiKey: "dummy-key"},
		Performance: configmodel.Performance{LibraryProcessingTimeout: 10 * time.Second},
		HTTPClient:  configmodel.HTTPClient{Timeout: 5 * time.Second},
		Logger:      configmodel.Logger{LogFilePath: "test.log", UseStdout: true, LogLevel: "info"},
		Processor:   configmodel.ProcessorConfig{ItemProcessor: configmodel.ItemProcessorConfig{RatingBuilder: configmodel.RatingBuilderConfig{Timeout: 1 * time.Second}}, LibraryProcessor: configmodel.LibraryProcessorConfig{DefaultTimeout: 1 * time.Second}},
	}
	initializer := NewServiceInitializer(s.logger, invalidConfig, s.ctx, s.workSemaphore)

	// Act
	err := initializer.InitializeServices()

	// Assert
	assert.Error(s.T(), err, "Expected an error due to missing Plex URL")
	s.T().Logf("Received error from PlexUrlMissing test: %v", err)
	assert.Contains(s.T(), err.Error(), "plex.url is required", "Error message should indicate Plex URL is required")
	assert.Empty(s.T(), initializer.GetMediaServices(), "Media services should be empty on config error")
}

func (s *ServiceInitializerSuite) TestInitializeServices_RatingServiceConfigError_TMDBApiKeyMissing() {
	// Arrange
	invalidConfig := &configmodel.Config{
		Plex: configmodel.Plex{
			Enabled: true, Url: "http://dummy", Token: "dummy",
		},
		TMDB: configmodel.TMDB{
			Enabled: true,
			ApiKey:  "",
		},
		Performance: configmodel.Performance{LibraryProcessingTimeout: 10 * time.Second},
		HTTPClient:  configmodel.HTTPClient{Timeout: 5 * time.Second},
		Logger:      configmodel.Logger{LogFilePath: "test.log", UseStdout: true, LogLevel: "info"},
		Processor:   configmodel.ProcessorConfig{ItemProcessor: configmodel.ItemProcessorConfig{RatingBuilder: configmodel.RatingBuilderConfig{Timeout: 1 * time.Second}}, LibraryProcessor: configmodel.LibraryProcessorConfig{DefaultTimeout: 1 * time.Second}},
	}

	initializer := NewServiceInitializer(s.logger, invalidConfig, s.ctx, s.workSemaphore)

	// Act
	err := initializer.InitializeServices()

	// Assert
	assert.Error(s.T(), err, "Expected an error due to missing TMDB API key")
	s.T().Logf("Received error from TMDBApiKeyMissing test: %v", err)
	assert.Contains(s.T(), err.Error(), "tmdb.api_key is required", "Error message should indicate TMDB API key is required")
	assert.Empty(s.T(), initializer.GetRatingPlatformServices(), "Rating platform services should be empty or only contain services initialized before the error")
}
