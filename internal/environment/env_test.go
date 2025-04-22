package env

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EnvTestSuite struct {
	suite.Suite
	originalEnvValue  string
	originalEnvSet    bool
	originalLogOutput io.Writer
}

// SetupTest runs before the tests in the suite, once per test method.
func (s *EnvTestSuite) SetupTest() {
	// Save original ENV state
	s.originalEnvValue, s.originalEnvSet = os.LookupEnv("ENV")
	// Unset ENV before each test method to ensure a clean slate.
	// Individual test cases or sub-tests will set it as needed.
	os.Unsetenv("ENV")

	// Suppress log output during tests
	s.originalLogOutput = log.Writer()
	log.SetOutput(io.Discard)
}

// TearDownTest runs after the tests in the suite, once per test method.
func (s *EnvTestSuite) TearDownTest() {
	// Restore original ENV state
	if s.originalEnvSet {
		os.Setenv("ENV", s.originalEnvValue)
	} else {
		os.Unsetenv("ENV")
	}

	// Restore log output
	log.SetOutput(s.originalLogOutput)
}

// TestLoad tests the Load function
func (s *EnvTestSuite) TestLoad() {
	s.Run("ENV_not_set_initially", func() {
		os.Unsetenv("ENV") // Ensure ENV is not set

		Load()

		assert.Equal(s.T(), "DEV", os.Getenv("ENV"), "ENV should be defaulted to DEV")
		os.Unsetenv("ENV") // Clean up for next sub-test
	})

	s.Run("ENV_already_set_to_PROD", func() {
		os.Setenv("ENV", "PROD") // Pre-set ENV

		Load()

		assert.Equal(s.T(), "PROD", os.Getenv("ENV"), "ENV should remain PROD")
		os.Unsetenv("ENV") // Clean up for next sub-test
	})

	s.Run("ENV_set_to_empty_string_initially", func() {
		os.Setenv("ENV", "") // Pre-set ENV to empty string

		Load()
		// os.Getenv("ENV") will be "" initially.
		// The Load function checks `if env == ""`, which will be true.
		// So, it should default to "DEV".
		assert.Equal(s.T(), "DEV", os.Getenv("ENV"), "ENV set to empty string should be defaulted to DEV")
		os.Unsetenv("ENV") // Clean up for next sub-test
	})
}

// TestGetEnvironment tests the GetEnvironment function
func (s *EnvTestSuite) TestGetEnvironment() {
	s.Run("ENV_not_set", func() {
		os.Unsetenv("ENV") // Ensure ENV is not set

		env := GetEnvironment()

		assert.Equal(s.T(), "DEV", env, "GetEnvironment should return DEV when ENV is not set")
	})

	s.Run("ENV_set_to_TESTING", func() {
		os.Setenv("ENV", "TESTING") // Pre-set ENV

		env := GetEnvironment()

		assert.Equal(s.T(), "TESTING", env, "GetEnvironment should return the set ENV value")
		os.Unsetenv("ENV") // Clean up for next sub-test
	})

	s.Run("ENV_set_to_empty_string", func() {
		os.Setenv("ENV", "") // Pre-set ENV to empty string

		env := GetEnvironment()
		// GetEnvironment checks `if env == ""`, which will be true.
		// So, it should return "DEV".
		assert.Equal(s.T(), "DEV", env, "GetEnvironment should return DEV when ENV is an empty string")
		os.Unsetenv("ENV") // Clean up for next sub-test
	})
}

// TestEnvTestSuite runs the test suite
func TestEnvTestSuite(t *testing.T) {
	suite.Run(t, new(EnvTestSuite))
}
