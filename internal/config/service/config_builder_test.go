package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	configModel "github.com/zepollabot/media-rating-overlay/internal/config/model"
)

type ConfigBuilderTestSuite struct {
	suite.Suite
	builder *ConfigBuilder
}

func (s *ConfigBuilderTestSuite) SetupTest() {
	s.builder = NewConfigBuilder()
}

func TestConfigBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigBuilderTestSuite))
}

func (s *ConfigBuilderTestSuite) TestNewConfigBuilder() {
	// Arrange
	builder := NewConfigBuilder()

	// Assert builder creation
	s.Assert().NotNil(builder, "NewConfigBuilder() should not return nil")

	// Act: Try to build the initial, empty config
	cfg, err := builder.Build()

	// Assert: Expect an error because an empty config has invalid zero-value fields by default
	s.Assert().Error(err, "Build() on a new builder should return an error due to invalid zero-value fields")
	// Check for one of the specific validation errors that would occur first
	s.Assert().Contains(err.Error(), "performance.library_processing_timeout must be positive", "Error message mismatch for new builder due to zero LibraryProcessingTimeout")
	s.Assert().Nil(cfg, "Config should be nil when Build() returns an error on a new builder")
}

func (s *ConfigBuilderTestSuite) TestWithDefaults() {
	// Act
	s.builder.WithDefaults()
	cfg, err := s.builder.Build()

	// Assert
	s.NoError(err, "Build() with defaults should not return an error")
	s.NotNil(cfg, "Config should not be nil after WithDefaults() and Build()")

	// Plex defaults
	s.False(cfg.Plex.Enabled, "Plex.Enabled should be false by default")

	// TMDB defaults
	s.False(cfg.TMDB.Enabled, "TMDB.Enabled should be false by default")
	s.Equal("en-US", cfg.TMDB.Language, "TMDB.Language should be 'en-US' by default")
	s.Equal("US", cfg.TMDB.Region, "TMDB.Region should be 'US' by default")

	// Performance defaults
	s.Equal(1, cfg.Performance.MaxThreads, "Performance.MaxThreads should be 1 by default (set based on CPU later)")
	s.Equal(600*time.Second, cfg.Performance.LibraryProcessingTimeout, "Performance.LibraryProcessingTimeout should be 600s by default")

	// HTTPClient defaults
	s.Equal(30*time.Second, cfg.HTTPClient.Timeout, "HTTPClient.Timeout should be 30s by default")
	s.Equal(3, cfg.HTTPClient.MaxRetries, "HTTPClient.MaxRetries should be 3 by default")

	// Logger defaults
	s.Equal("info", cfg.Logger.LogLevel, "Logger.LogLevel should be 'info' by default")
	s.Equal(100, cfg.Logger.MaxSize, "Logger.MaxSize should be 100 by default")
	s.Equal(5, cfg.Logger.MaxBackups, "Logger.MaxBackups should be 5 by default")
	s.Equal(30, cfg.Logger.MaxAge, "Logger.MaxAge should be 30 by default")
	s.True(cfg.Logger.Compress, "Logger.Compress should be true by default")
	s.Equal("media-rating-overlay", cfg.Logger.ServiceName, "Logger.ServiceName should be 'media-rating-overlay' by default")
	s.Equal("1.0.0", cfg.Logger.ServiceVersion, "Logger.ServiceVersion should be '1.0.0' by default")
	s.True(cfg.Logger.UseJSON, "Logger.UseJSON should be true by default")
	s.True(cfg.Logger.UseStdout, "Logger.UseStdout should be true by default")

	// Processor defaults
	s.Equal(30*time.Second, cfg.Processor.ItemProcessor.RatingBuilder.Timeout, "Processor.ItemProcessor.RatingBuilder.Timeout should be 30s by default")
	s.Equal(600*time.Second, cfg.Processor.LibraryProcessor.DefaultTimeout, "Processor.LibraryProcessor.DefaultTimeout should be 600s by default")
}

func (s *ConfigBuilderTestSuite) TestWithPlex() {
	// Arrange
	s.builder.WithDefaults() // Apply defaults first
	plexConfig := configModel.Plex{
		Enabled: true,
		Url:     "http://localhost:32400",
		Token:   "test-token",
	}

	// Act
	s.builder.WithPlex(plexConfig)
	cfg, err := s.builder.Build()

	// Assert
	s.NoError(err)
	s.NotNil(cfg)
	s.Equal(plexConfig, cfg.Plex)
}

func (s *ConfigBuilderTestSuite) TestWithTMDB() {
	// Arrange
	s.builder.WithDefaults() // Apply defaults first
	tmdbConfig := configModel.TMDB{
		Enabled:  true,
		ApiKey:   "test-api-key",
		Language: "es-ES", // Custom language to ensure it's set over default
		Region:   "ES",    // Custom region
	}

	// Act
	s.builder.WithTMDB(tmdbConfig)
	cfg, err := s.builder.Build()

	// Assert
	s.NoError(err)
	s.NotNil(cfg)
	s.Equal(tmdbConfig, cfg.TMDB)
}

func (s *ConfigBuilderTestSuite) TestWithPerformance() {
	// Arrange
	s.builder.WithDefaults() // Apply defaults first
	performanceConfig := configModel.Performance{
		MaxThreads:               4,
		LibraryProcessingTimeout: 60 * time.Second,
	}

	// Act
	s.builder.WithPerformance(performanceConfig)
	cfg, err := s.builder.Build()

	// Assert
	s.NoError(err)
	s.NotNil(cfg)
	s.Equal(performanceConfig, cfg.Performance)
}

func (s *ConfigBuilderTestSuite) TestWithHTTPClient() {
	// Arrange
	s.builder.WithDefaults() // Apply defaults first
	httpClientConfig := configModel.HTTPClient{
		Timeout:    15 * time.Second,
		MaxRetries: 5,
	}

	// Act
	s.builder.WithHTTPClient(httpClientConfig)
	cfg, err := s.builder.Build()

	// Assert
	s.NoError(err)
	s.NotNil(cfg)
	s.Equal(httpClientConfig, cfg.HTTPClient)
}

func (s *ConfigBuilderTestSuite) TestWithLogger() {
	// Arrange
	s.builder.WithDefaults() // Apply defaults first
	loggerConfig := configModel.Logger{
		LogLevel: "debug",
		UseJSON:  false,
		// Ensure all fields are explicitly set for comparison if they differ from zero values after WithDefaults + WithLogger
		MaxSize:        50,               // Custom value
		MaxBackups:     3,                // Custom value
		MaxAge:         15,               // Custom value
		Compress:       false,            // Custom value
		ServiceName:    "custom-service", // Custom value
		ServiceVersion: "2.0.0",          // Custom value
		UseStdout:      false,            // Custom value
	}

	// Act
	s.builder.WithLogger(loggerConfig)
	cfg, err := s.builder.Build()

	// Assert
	s.NoError(err)
	s.NotNil(cfg)
	s.Equal(loggerConfig, cfg.Logger)
}

func (s *ConfigBuilderTestSuite) TestWithProcessor() {
	// Arrange
	s.builder.WithDefaults() // Apply defaults first
	processorConfig := configModel.ProcessorConfig{
		ItemProcessor: configModel.ItemProcessorConfig{
			RatingBuilder: configModel.RatingBuilderConfig{
				Timeout: 15 * time.Second,
			},
		},
		LibraryProcessor: configModel.LibraryProcessorConfig{
			DefaultTimeout: 60 * time.Second,
		},
	}

	// Act
	s.builder.WithProcessor(processorConfig)
	cfg, err := s.builder.Build()

	// Assert
	s.NoError(err)
	s.NotNil(cfg)
	s.Equal(processorConfig, cfg.Processor)
}

func (s *ConfigBuilderTestSuite) TestBuild_ValidConfig() {
	// Arrange
	s.builder.WithDefaults().
		WithPlex(configModel.Plex{Enabled: true, Url: "http://plex.local", Token: "plex-token"}).
		WithTMDB(configModel.TMDB{Enabled: true, ApiKey: "tmdb-key"})

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.NoError(err, "Build() with valid config should not return an error")
	s.NotNil(cfg, "Config should not be nil for a valid build")
}

func (s *ConfigBuilderTestSuite) TestBuild_PlexEnabledNoUrl() {
	// Arrange
	s.builder.WithDefaults().WithPlex(configModel.Plex{Enabled: true, Token: "some-token"})

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for Plex enabled with no URL")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "plex.url is required when plex is enabled", "Error message mismatch")
}

func (s *ConfigBuilderTestSuite) TestBuild_PlexEnabledNoToken() {
	// Arrange
	s.builder.WithDefaults().WithPlex(configModel.Plex{Enabled: true, Url: "http://plex.local"})

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for Plex enabled with no Token")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "plex.token is required when plex is enabled", "Error message mismatch")
}

func (s *ConfigBuilderTestSuite) TestBuild_TMDBEnabledNoApiKey() {
	// Arrange
	s.builder.WithDefaults().WithTMDB(configModel.TMDB{Enabled: true})

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for TMDB enabled with no API key")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "tmdb.api_key is required when tmdb is enabled", "Error message mismatch")
}

func (s *ConfigBuilderTestSuite) TestBuild_InvalidPerformanceMaxThreads() {
	// Arrange
	s.builder.WithDefaults().WithPerformance(configModel.Performance{MaxThreads: -1, LibraryProcessingTimeout: 10 * time.Second})

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for negative MaxThreads")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "performance.max_threads must be non-negative", "Error message mismatch")
}

func (s *ConfigBuilderTestSuite) TestBuild_InvalidPerformanceLibraryProcessingTimeout() {
	// Arrange
	s.builder.WithDefaults().WithPerformance(configModel.Performance{MaxThreads: 1, LibraryProcessingTimeout: 0})

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for non-positive LibraryProcessingTimeout")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "performance.library_processing_timeout must be positive", "Error message mismatch")
}

func (s *ConfigBuilderTestSuite) TestBuild_InvalidHTTPClientTimeout() {
	// Arrange
	s.builder.WithDefaults().WithHTTPClient(configModel.HTTPClient{Timeout: 0, MaxRetries: 1})

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for non-positive HTTPClient Timeout")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "http_client.timeout must be positive", "Error message mismatch")
}

func (s *ConfigBuilderTestSuite) TestBuild_InvalidHTTPClientMaxRetries() {
	// Arrange
	s.builder.WithDefaults().WithHTTPClient(configModel.HTTPClient{Timeout: 10 * time.Second, MaxRetries: -1})

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for negative HTTPClient MaxRetries")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "http_client.max_retries must be non-negative", "Error message mismatch")
}

func (s *ConfigBuilderTestSuite) TestBuild_InvalidProcessorItemRatingBuilderTimeout() {
	// Arrange
	procConf := configModel.ProcessorConfig{
		ItemProcessor: configModel.ItemProcessorConfig{
			RatingBuilder: configModel.RatingBuilderConfig{
				Timeout: 0,
			},
		},
		LibraryProcessor: configModel.LibraryProcessorConfig{
			DefaultTimeout: 1 * time.Second,
		},
	}
	s.builder.WithDefaults().WithProcessor(procConf)

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for non-positive ItemProcessor.RatingBuilder.Timeout")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "processor.item_processor.rating_builder.timeout must be positive", "Error message mismatch")
}

func (s *ConfigBuilderTestSuite) TestBuild_InvalidProcessorLibraryDefaultTimeout() {
	// Arrange
	procConf := configModel.ProcessorConfig{
		ItemProcessor: configModel.ItemProcessorConfig{
			RatingBuilder: configModel.RatingBuilderConfig{
				Timeout: 1 * time.Second,
			},
		},
		LibraryProcessor: configModel.LibraryProcessorConfig{
			DefaultTimeout: 0,
		},
	}
	s.builder.WithDefaults().WithProcessor(procConf)

	// Act
	cfg, err := s.builder.Build()

	// Assert
	s.Error(err, "Build() should return an error for non-positive LibraryProcessor.DefaultTimeout")
	s.Nil(cfg, "Config should be nil on validation error")
	s.Contains(err.Error(), "processor.library_processor.default_timeout must be positive", "Error message mismatch")
}
