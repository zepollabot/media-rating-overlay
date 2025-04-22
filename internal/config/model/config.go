package config

import (
	"fmt"
)

// Config holds the application configuration
type Config struct {
	Plex        Plex            `yaml:"plex"`
	TMDB        TMDB            `yaml:"tmdb"`
	Performance Performance     `yaml:"performance"`
	HTTPClient  HTTPClient      `yaml:"http_client"`
	Logger      Logger          `yaml:"logger"`
	Processor   ProcessorConfig `yaml:"processor"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	config := &Config{}
	config.Plex = *DefaultPlex()
	config.TMDB = *DefaultTMDB()
	config.Performance = *DefaultPerformance()
	config.HTTPClient = *DefaultHTTPClient()
	config.Logger = *DefaultLogger()
	config.Processor = *DefaultProcessorConfig()
	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Plex.Validate(); err != nil {
		return fmt.Errorf("plex config: %w", err)
	}
	if err := c.TMDB.Validate(); err != nil {
		return fmt.Errorf("tmdb config: %w", err)
	}
	if err := c.Performance.Validate(); err != nil {
		return fmt.Errorf("performance config: %w", err)
	}
	if err := c.HTTPClient.Validate(); err != nil {
		return fmt.Errorf("http_client config: %w", err)
	}
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger config: %w", err)
	}
	if err := c.Processor.Validate(); err != nil {
		return fmt.Errorf("processor config: %w", err)
	}
	return nil
}
