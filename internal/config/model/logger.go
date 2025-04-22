package config

import "fmt"

// Logger configuration for the application
type Logger struct {
	LogFilePath    string `yaml:"log_file_path"`
	LogLevel       string `yaml:"log_level"`
	MaxSize        int    `yaml:"max_size"`
	MaxBackups     int    `yaml:"max_backups"`
	MaxAge         int    `yaml:"max_age"`
	Compress       bool   `yaml:"compress"`
	ServiceName    string `yaml:"service_name"`
	ServiceVersion string `yaml:"service_version"`
	UseJSON        bool   `yaml:"use_json"`
	UseStdout      bool   `yaml:"use_stdout"`
}

func DefaultLogger() *Logger {
	return &Logger{
		LogFilePath:    "logs/media-rating-overlay.log",
		LogLevel:       "info",
		MaxSize:        100, // 100 MB
		MaxBackups:     5,
		MaxAge:         30, // 30 days
		Compress:       true,
		ServiceName:    "media-rating-overlay",
		ServiceVersion: "1.0.0",
		UseJSON:        true,
		UseStdout:      true,
	}
}

// Validate validates the Logger configuration
func (c *Logger) Validate() error {
	if c.LogFilePath == "" {
		return fmt.Errorf("logger.log_file_path is required")
	}
	if c.MaxSize <= 0 {
		return fmt.Errorf("logger.max_size must be positive")
	}
	if c.MaxBackups < 0 {
		return fmt.Errorf("logger.max_backups must be non-negative")
	}
	if c.MaxAge < 0 {
		return fmt.Errorf("logger.max_age must be non-negative")
	}
	if c.ServiceName == "" {
		return fmt.Errorf("logger.service_name is required")
	}
	if c.ServiceVersion == "" {
		return fmt.Errorf("logger.service_version is required")
	}
	return nil
}
