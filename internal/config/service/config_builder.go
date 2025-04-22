package config

import (
	"fmt"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
)

// ConfigBuilder builds configuration with defaults and validation
type ConfigBuilder struct {
	config *config.Config
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &config.Config{},
	}
}

// WithDefaults sets default values for the configuration
func (b *ConfigBuilder) WithDefaults() *ConfigBuilder {
	b.config.Plex = config.Plex{
		Enabled: false,
	}
	b.config.TMDB = config.TMDB{
		Enabled:  false,
		Language: "en-US",
		Region:   "US",
	}
	b.config.Performance = *config.DefaultPerformance()
	b.config.HTTPClient = *config.DefaultHTTPClient()
	b.config.Logger = *config.DefaultLogger()
	b.config.Processor = *config.DefaultProcessorConfig()
	return b
}

// WithPlex sets Plex configuration
func (b *ConfigBuilder) WithPlex(plex config.Plex) *ConfigBuilder {
	b.config.Plex = plex
	return b
}

// WithTMDB sets TMDB configuration
func (b *ConfigBuilder) WithTMDB(tmdb config.TMDB) *ConfigBuilder {
	b.config.TMDB = tmdb
	return b
}

// WithPerformance sets performance configuration
func (b *ConfigBuilder) WithPerformance(performance config.Performance) *ConfigBuilder {
	b.config.Performance = performance
	return b
}

// WithHTTPClient sets HTTP client configuration
func (b *ConfigBuilder) WithHTTPClient(httpClient config.HTTPClient) *ConfigBuilder {
	b.config.HTTPClient = httpClient
	return b
}

// WithLogger sets logger configuration
func (b *ConfigBuilder) WithLogger(logger config.Logger) *ConfigBuilder {
	b.config.Logger = logger
	return b
}

// WithProcessor sets processor configuration
func (b *ConfigBuilder) WithProcessor(processor config.ProcessorConfig) *ConfigBuilder {
	b.config.Processor = processor
	return b
}

// Build validates and returns the configuration
func (b *ConfigBuilder) Build() (*config.Config, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// validate performs configuration validation
func (b *ConfigBuilder) validate() error {
	if b.config.Plex.Enabled {
		if b.config.Plex.Url == "" {
			return fmt.Errorf("plex.url is required when plex is enabled")
		}
		if b.config.Plex.Token == "" {
			return fmt.Errorf("plex.token is required when plex is enabled")
		}
	}

	if b.config.TMDB.Enabled {
		if b.config.TMDB.ApiKey == "" {
			return fmt.Errorf("tmdb.api_key is required when tmdb is enabled")
		}
	}

	if b.config.Performance.MaxThreads < 0 {
		return fmt.Errorf("performance.max_threads must be non-negative")
	}
	if b.config.Performance.LibraryProcessingTimeout <= 0 {
		return fmt.Errorf("performance.library_processing_timeout must be positive")
	}

	if b.config.HTTPClient.Timeout <= 0 {
		return fmt.Errorf("http_client.timeout must be positive")
	}
	if b.config.HTTPClient.MaxRetries < 0 {
		return fmt.Errorf("http_client.max_retries must be non-negative")
	}

	if b.config.Processor.ItemProcessor.RatingBuilder.Timeout <= 0 {
		return fmt.Errorf("processor.item_processor.rating_builder.timeout must be positive")
	}
	if b.config.Processor.LibraryProcessor.DefaultTimeout <= 0 {
		return fmt.Errorf("processor.library_processor.default_timeout must be positive")
	}

	return nil
}
