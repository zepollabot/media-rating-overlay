package config

import (
	"fmt"
	"time"
)

// ProcessorConfig holds configuration for processors
type ProcessorConfig struct {
	// ItemProcessor configuration
	ItemProcessor ItemProcessorConfig `yaml:"item_processor"`

	// LibraryProcessor configuration
	LibraryProcessor LibraryProcessorConfig `yaml:"library_processor"`
}

type RatingBuilderConfig struct {
	Timeout time.Duration `yaml:"timeout"`
}

type ItemProcessorConfig struct {
	RatingBuilder RatingBuilderConfig `yaml:"rating_builder"`
}

type LibraryProcessorConfig struct {
	DefaultTimeout time.Duration `yaml:"default_timeout"`
}

// DefaultProcessorConfig returns a default processor configuration
func DefaultProcessorConfig() *ProcessorConfig {
	config := &ProcessorConfig{}

	// Item processor defaults
	config.ItemProcessor.RatingBuilder.Timeout = 30 * time.Second

	// Library processor defaults
	config.LibraryProcessor.DefaultTimeout = 600 * time.Second // 10 minutes

	return config
}

// Validate validates the processor configuration
func (c *ProcessorConfig) Validate() error {
	if c.ItemProcessor.RatingBuilder.Timeout <= 0 {
		return fmt.Errorf("rating_builder.timeout must be greater than 0")
	}
	if c.LibraryProcessor.DefaultTimeout <= 0 {
		return fmt.Errorf("default_timeout must be greater than 0")
	}
	return nil
}
