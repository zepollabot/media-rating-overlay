package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	models "github.com/zepollabot/media-rating-overlay/internal/config/model"
)

const (
	// Default file and directory names for configuration
	DefaultConfigFileName       = "config.yaml"
	DefaultEnvConfigDir         = "configs"
	DefaultEnvConfigFilePattern = "config.env.%s.yaml"

	// Environment names
	envDevelopment = "DEV"
	envProduction  = "PROD"

	// Log levels
	logLevelDebug = "debug"
)

type ConfigService struct {
	ConfigDirPath        string // Base directory for configuration files
	DefaultConfigName    string // Name of the default/base configuration file
	EnvConfigFilePattern string // Pattern for environment-specific config files
	Env                  string // Current operating environment (e.g., "DEV", "PROD")
}

// NewConfigService creates a new ConfigService with specified configurations.
// This constructor allows for custom paths and patterns, useful for testing or specific setups.
// If env is an empty string, it defaults to envDevelopment ("DEV").
func NewConfigService(configDirPath, defaultConfigName, envConfigFilePattern, env string) *ConfigService {
	effectiveEnv := env
	if effectiveEnv == "" {
		effectiveEnv = envDevelopment
	}
	return &ConfigService{
		ConfigDirPath:        configDirPath,
		DefaultConfigName:    defaultConfigName,
		EnvConfigFilePattern: envConfigFilePattern,
		Env:                  effectiveEnv,
	}
}

// NewDefaultConfigService creates a ConfigService using default file and directory names.
// It retrieves the environment from the "ENV" environment variable, defaulting to envDevelopment ("DEV")
// if the environment variable is not set.
func NewDefaultConfigService() *ConfigService {
	env := os.Getenv("ENV")
	// effectiveEnv will be set to envDevelopment by NewConfigService if env is empty.
	return NewConfigService(DefaultEnvConfigDir, DefaultConfigFileName, DefaultEnvConfigFilePattern, env)
}

// Load loads the configuration based on the current environment
func (c *ConfigService) Load() (*models.Config, error) {
	// Load base configuration (environment-independent)
	baseConfigPath := filepath.Join(c.ConfigDirPath, c.DefaultConfigName)
	baseConfig, err := c.loadFromFile(baseConfigPath)

	if err != nil {
		return nil, fmt.Errorf("error loading base configuration from %s: %w", baseConfigPath, err)
	}

	currentConfig := baseConfig

	// Try to load environment-specific config
	envConfigFileName := fmt.Sprintf(c.EnvConfigFilePattern, strings.ToLower(c.Env))
	envConfigPath := filepath.Join(c.ConfigDirPath, envConfigFileName)
	envConfig, err := c.loadFromFile(envConfigPath)

	if err == nil {
		// Environment-specific config loaded, merge it with base config
		mergedConfig, mergeErr := c.mergeConfigs(baseConfig, envConfig)
		if mergeErr != nil {
			return nil, fmt.Errorf("error merging configurations: %w", mergeErr)
		}
		currentConfig = mergedConfig
	} else if !os.IsNotExist(err) {
		// An error other than "file not found" occurred when loading env-specific config
		return nil, fmt.Errorf("error loading environment-specific configuration from %s: %w", envConfigPath, err)
	}
	// If os.IsNotExist(err) is true, it means no environment-specific config file was found,
	// so we proceed with currentConfig (which is baseConfig).

	// Validate the final configuration for the current environment
	if err := c.validateEnvironmentSpecificRules(currentConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration for environment %s: %w", c.Env, err)
	}

	return currentConfig, nil
}

// mergeConfigs merges environment-specific config with base config
func (c *ConfigService) mergeConfigs(base, env *models.Config) (*models.Config, error) {
	// Create a copy of the base config
	merged := *base

	// Override with environment-specific values if they exist and are non-empty/non-zero

	// Plex
	if env.Plex.Enabled { // Gate: Is the Plex config in 'env' to be considered?
		merged.Plex.Enabled = true // If env.Plex is enabled, the merged Plex is enabled
		if env.Plex.Url != "" {
			merged.Plex.Url = env.Plex.Url
		}
		if env.Plex.Token != "" { // Only override if env token is non-empty
			merged.Plex.Token = env.Plex.Token
		}
	}

	// TMDB
	if env.TMDB.Enabled { // Gate
		merged.TMDB.Enabled = true
		if env.TMDB.ApiKey != "" {
			merged.TMDB.ApiKey = env.TMDB.ApiKey
		}
	}

	// Performance
	if env.Performance.MaxThreads != 0 || env.Performance.LibraryProcessingTimeout != 0 {
		if env.Performance.MaxThreads != 0 {
			merged.Performance.MaxThreads = env.Performance.MaxThreads
		}
		if env.Performance.LibraryProcessingTimeout != 0 {
			merged.Performance.LibraryProcessingTimeout = env.Performance.LibraryProcessingTimeout
		}
	}

	// HTTPClient
	if env.HTTPClient.Timeout > 0 { // Gate
		merged.HTTPClient.Timeout = env.HTTPClient.Timeout
	}

	// Logger
	if env.Logger.LogFilePath != "" || env.Logger.LogLevel != "" {
		if env.Logger.LogFilePath != "" {
			merged.Logger.LogFilePath = env.Logger.LogFilePath
		}
		if env.Logger.LogLevel != "" {
			merged.Logger.LogLevel = env.Logger.LogLevel
		}
	}

	// Processor
	// Using a more general gate for Processor: if any of its configurable parts have non-zero values from env.
	if (env.Processor.ItemProcessor.RatingBuilder.Timeout != 0) ||
		(env.Processor.LibraryProcessor.DefaultTimeout != 0) {

		// ItemProcessor fields
		if env.Processor.ItemProcessor.RatingBuilder.Timeout != 0 {
			merged.Processor.ItemProcessor.RatingBuilder.Timeout = env.Processor.ItemProcessor.RatingBuilder.Timeout
		}

		// LibraryProcessor fields
		if env.Processor.LibraryProcessor.DefaultTimeout != 0 {
			merged.Processor.LibraryProcessor.DefaultTimeout = env.Processor.LibraryProcessor.DefaultTimeout
		}
	}

	// Validate the merged config
	if err := merged.Validate(); err != nil {
		return nil, fmt.Errorf("invalid merged configuration: %w", err)
	}

	return &merged, nil
}

// loadFromFile loads configuration from a specific file
func (c *ConfigService) loadFromFile(path string) (*models.Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config models.Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	// Use the builder to ensure defaults and validation
	builder := NewConfigBuilder().WithDefaults()
	if config.Plex.Enabled {
		builder.WithPlex(config.Plex)
	}
	if config.TMDB.Enabled {
		builder.WithTMDB(config.TMDB)
	}
	if config.Performance.MaxThreads != 0 {
		builder.WithPerformance(config.Performance)
	}
	if config.HTTPClient.Timeout > 0 {
		builder.WithHTTPClient(config.HTTPClient)
	}
	if config.Logger.LogFilePath != "" {
		builder.WithLogger(config.Logger)
	}
	// Check if any part of Processor config is set before calling WithProcessor
	if config.Processor.ItemProcessor.RatingBuilder.Timeout > 0 ||
		config.Processor.LibraryProcessor.DefaultTimeout > 0 {
		builder.WithProcessor(config.Processor)
	}

	return builder.Build()
}

// validateEnvironmentSpecificRules validates environment-specific rules
// This is separate from the general configuration validation which is handled by the builder
func (c *ConfigService) validateEnvironmentSpecificRules(config *models.Config) error {
	if c.Env == envProduction {
		// Ensure production has secure settings
		if config.Logger.LogLevel == logLevelDebug {
			return fmt.Errorf("debug logging is not allowed in production")
		}
	}
	return nil
}
