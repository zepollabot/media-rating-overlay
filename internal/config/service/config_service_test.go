package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	configModels "github.com/zepollabot/media-rating-overlay/internal/config/model"
)

const (
	// testDir is the name of the temporary directory for test configuration files.
	testDir = "test_temp_configs"
	// baseConfigTestFileName is the name of the base configuration file for tests.
	baseConfigTestFileName = "config.yaml"
	// envConfigTestFilePattern is the pattern for environment-specific config files in tests.
	envConfigTestFilePattern = "config.env.%s.yaml"
	// testEnvironment is a specific environment name used for testing.
	testEnvironment = "unittest"
)

// ConfigServiceTestSuite defines the test suite for the
type ConfigServiceTestSuite struct {
	suite.Suite
	tempConfigDirPath string         // Path to the temporary directory holding config files for a test.
	service           *ConfigService // Instance of the service under test.
}

// SetupSuite runs once before all tests in this suite.
// It creates the main temporary directory for configuration files.
func (s *ConfigServiceTestSuite) SetupSuite() {
	// Determine path relative to the package being tested or project root
	// This often resolves to being next to the package under test.
	cwd, err := os.Getwd()
	s.Require().NoError(err, "Failed to get current working directory")
	s.tempConfigDirPath = filepath.Join(cwd, testDir)

	// Ensure the directory does not exist from a previous failed run
	_ = os.RemoveAll(s.tempConfigDirPath)

	err = os.MkdirAll(s.tempConfigDirPath, 0750)
	s.Require().NoError(err, "Failed to create temporary config directory for suite")
}

// TearDownSuite runs once after all tests in this suite are finished.
// It cleans up the main temporary directory.
func (s *ConfigServiceTestSuite) TearDownSuite() {
	err := os.RemoveAll(s.tempConfigDirPath)
	s.Require().NoError(err, "Failed to remove temporary config directory after suite")
}

// SetupTest runs before each test method.
// It ensures the temporary directory is clean for the upcoming test.
func (s *ConfigServiceTestSuite) SetupTest() {
	// Clean contents of the temp directory, but not the directory itself
	dirEntries, err := os.ReadDir(s.tempConfigDirPath)
	s.Require().NoError(err, "Failed to read temp config directory before test")
	for _, entry := range dirEntries {
		err = os.RemoveAll(filepath.Join(s.tempConfigDirPath, entry.Name()))
		s.Require().NoError(err, "Failed to remove item %s from temp config dir", entry.Name())
	}
}

// createTestConfigFile is a helper to create a YAML configuration file with the given content.
func (s *ConfigServiceTestSuite) createTestConfigFile(fileName string, content *configModels.Config) string {
	filePath := filepath.Join(s.tempConfigDirPath, fileName)
	data, err := yaml.Marshal(content)
	s.Require().NoError(err, "Failed to marshal test config content for %s", fileName)
	err = os.WriteFile(filePath, data, 0640)
	s.Require().NoError(err, "Failed to write test config file %s", filePath)
	return filePath
}

// TestLoad_HappyPath_WithBaseAndEnvFiles tests the Load method
// when both a base configuration file and an environment-specific one exist.
// It checks if the configurations are merged correctly, with environment-specific
// values taking precedence.
func (s *ConfigServiceTestSuite) TestLoad_WithBaseAndEnvFiles() {
	// Arrange
	// Define base configuration content
	// Note: the current `loadFromFile` logic is `NewConfigBuilder().WithDefaults()`
	baseContent := configModels.DefaultConfig()
	s.createTestConfigFile(baseConfigTestFileName, baseContent)

	// Define environment-specific configuration content (overriding some values)
	envSpecificContent := &configModels.Config{
		Plex: configModels.Plex{
			Enabled: true, // Must be true to be picked by loadFromFile and mergeConfigs
			Url:     "http://env-plex.com",
			Token:   "env-token",
		},
		Logger: configModels.Logger{
			LogLevel:       "debug",
			LogFilePath:    "/var/log/env_test.log", // Condition: Logger.LogFilePath != ""
			MaxSize:        100,                     // 100 MB
			MaxBackups:     5,
			MaxAge:         30, // 30 days
			Compress:       true,
			ServiceName:    "media-rating-overlay",
			ServiceVersion: "1.0.0",
			UseJSON:        true,
			UseStdout:      true,
		},
		// TMDB is not specified in env config: expecting it to be taken from baseContent.
		// Performance is not specified: expecting it from baseContent.
		// HTTPClient is not specified: expecting it from baseContent (as base had Timeout > 0)
		// Processor is not specified: expecting it from baseContent
	}
	envConfigFileName := fmt.Sprintf(envConfigTestFilePattern, testEnvironment)
	s.createTestConfigFile(envConfigFileName, envSpecificContent)

	// Initialize the service to use these temporary files
	s.service = NewConfigService(s.tempConfigDirPath, baseConfigTestFileName, envConfigTestFilePattern, testEnvironment)

	// Act
	loadedConfig, err := s.service.Load()

	fmt.Printf("loadedConfig: %+v\n", loadedConfig)

	// Assert
	s.Require().NoError(err, "Load() should not return an error in this happy path scenario")
	s.Require().NotNil(loadedConfig, "Loaded configuration should not be nil")

	// 1. Plex: Values from env-specific should override base. Token should be empty as it's not in env-specific Plex.
	s.True(loadedConfig.Plex.Enabled, "Plex.Enabled should be true (from env)")
	s.Equal("http://env-plex.com", loadedConfig.Plex.Url, "Plex.URL should be from env")
	s.Equal("env-token", loadedConfig.Plex.Token, "Plex.Token should be from env")

	// 2. TMDB: Should be taken entirely from base, as not specified in env.
	s.False(loadedConfig.TMDB.Enabled, "TMDB.Enabled should be false (from base)")
	s.Equal("en-US", loadedConfig.TMDB.Language, "TMDB.Language should be from base")
	s.Equal("US", loadedConfig.TMDB.Region, "TMDB.Region should be from base")
	s.Equal(loadedConfig.TMDB.ApiKey, "", "TMDB.APIKey should be empty (from base)")

	// 3. Logger: Values from env-specific should override base.
	s.Equal("debug", loadedConfig.Logger.LogLevel, "Logger.LogLevel should be from env")
	s.Equal("/var/log/env_test.log", loadedConfig.Logger.LogFilePath, "Logger.LogFilePath should be from env")

	// 4. Performance: Should be taken from base.
	s.Equal(baseContent.Performance.MaxThreads, loadedConfig.Performance.MaxThreads, "Performance.MaxThreads should be from base")

	// 5. HTTPClient: Should be taken from base (since base met the load condition).
	//    If base didn't meet the condition, it would be default.
	s.Equal(baseContent.HTTPClient.Timeout, loadedConfig.HTTPClient.Timeout, "HTTPClient.Timeout should be from base")

	// 6. Processor: Should be taken from base.
	s.Equal(baseContent.Processor.ItemProcessor.RatingBuilder.Timeout, loadedConfig.Processor.ItemProcessor.RatingBuilder.Timeout, "Processor.ItemProcessor.RatingBuilder.Timeout should be from base")
	s.Equal(baseContent.Processor.LibraryProcessor.DefaultTimeout, loadedConfig.Processor.LibraryProcessor.DefaultTimeout, "Processor.LibraryProcessor.DefaultTimeout should be from base")

	// 7. Environment specific rules validation (debug logging is allowed for non-PROD env "unittest")
	// This is implicitly checked by `s.Require().NoError(err)` above, as `validateEnvironmentSpecificRules`
	// would return an error if "debug" was not allowed for `testEnvironment`.
}

// TestLoad_OnlyBaseConfigExists tests the Load method when only a base configuration file exists.
func (s *ConfigServiceTestSuite) TestLoad_OnlyBaseConfigExists() {
	// Arrange
	baseContent := &configModels.Config{
		Plex: configModels.Plex{
			Enabled: true,
			Url:     "http://plex-base-only.com",
			Token:   "plex-base-only-token",
		},
		Logger: configModels.Logger{
			LogLevel:    "info",
			LogFilePath: "/var/log/base_only.log",
		},
	}
	s.createTestConfigFile(baseConfigTestFileName, baseContent)

	// Env-specific file does NOT exist for this test.

	s.service = NewConfigService(s.tempConfigDirPath, baseConfigTestFileName, envConfigTestFilePattern, testEnvironment)

	// Act
	loadedConfig, err := s.service.Load()

	// Assert
	s.Require().NoError(err, "Load() should not return an error when only base config exists")
	s.Require().NotNil(loadedConfig, "Loaded configuration should not be nil")

	s.Equal(baseContent.Plex.Url, loadedConfig.Plex.Url)
	s.Equal(baseContent.Plex.Token, loadedConfig.Plex.Token)
	s.Equal(baseContent.Logger.LogLevel, loadedConfig.Logger.LogLevel)
	s.Equal(baseContent.Logger.LogFilePath, loadedConfig.Logger.LogFilePath)
}

// TestLoad_BaseConfigMissing_Error tests the Load method when the base configuration file is missing.
func (s *ConfigServiceTestSuite) TestLoad_BaseConfigMissing_Error() {
	// Arrange
	// Do NOT create baseConfigTestFileName

	s.service = NewConfigService(s.tempConfigDirPath, baseConfigTestFileName, envConfigTestFilePattern, testEnvironment)

	// Act
	loadedConfig, err := s.service.Load()

	// Assert
	s.Require().Error(err, "Load() should return an error if base config is missing")
	// Check that the error is about the base configuration file path.
	// The error message includes the path, so we can check for its presence.
	expectedBaseConfigPath := filepath.Join(s.tempConfigDirPath, baseConfigTestFileName)
	s.Contains(err.Error(), expectedBaseConfigPath, "Error message should contain the base config path")
	s.Nil(loadedConfig, "Loaded configuration should be nil when base config is missing")
}

// TestLoad_ProdEnv_DebugLogging_Error tests the Load method for PROD environment validation failure (debug logging).
func (s *ConfigServiceTestSuite) TestLoad_ProdEnv_DebugLogging_Error() {
	// Arrange
	baseContent := &configModels.Config{
		Logger: configModels.Logger{
			LogLevel:    "debug",                        // Set to debug using string literal
			LogFilePath: "/var/log/prod_debug_test.log", // Needs to be non-empty for logger to be considered by builder
		},
	}
	s.createTestConfigFile(baseConfigTestFileName, baseContent)

	// Initialize service for PROD environment
	s.service = NewConfigService(s.tempConfigDirPath, baseConfigTestFileName, envConfigTestFilePattern, "PROD") // Use string literal for PROD

	// Act
	loadedConfig, err := s.service.Load()

	// Assert
	s.Require().Error(err, "Load() should return an error for debug logging in PROD")
	s.Contains(err.Error(), "debug logging is not allowed in production", "Error message should indicate debug logging is not allowed")
	s.Nil(loadedConfig, "Loaded configuration should be nil on validation failure")
}

// TestMain is the entry point for running the tests in this suite.
func TestConfigServiceSuite(t *testing.T) {
	suite.Run(t, new(ConfigServiceTestSuite))
}
