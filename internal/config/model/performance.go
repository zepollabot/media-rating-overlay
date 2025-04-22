package config

import (
	"fmt"
	"time"
)

// Performance represents performance-related configuration
type Performance struct {
	MaxThreads               int           `yaml:"max_threads"`
	LibraryProcessingTimeout time.Duration `yaml:"library_processing_timeout"`
}

func DefaultPerformance() *Performance {
	return &Performance{
		MaxThreads:               1,                 // Will be set based on CPU count
		LibraryProcessingTimeout: 600 * time.Second, // 10 minutes
	}
}

// Validate validates the Performance configuration
func (c *Performance) Validate() error {
	if c.MaxThreads < 0 {
		return fmt.Errorf("performance.max_threads must be non-negative")
	}
	if c.LibraryProcessingTimeout <= 0 {
		return fmt.Errorf("performance.library_processing_timeout must be positive")
	}
	return nil
}
