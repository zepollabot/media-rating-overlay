package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	clientmocks "github.com/zepollabot/media-rating-overlay/internal/httpclient/mocks"
)

const testLogFilePath = "test_log.txt"

type LoggingTestSuite struct {
	suite.Suite
	mockClient *clientmocks.HTTPClient
}

func (s *LoggingTestSuite) SetupTest() {
	s.mockClient = clientmocks.NewHTTPClient(s.T())
}

func (s *LoggingTestSuite) TearDownTest() {
	s.mockClient.AssertExpectations(s.T())
	_ = os.Remove(testLogFilePath) // Clean up the test log file
}

func TestLoggingTestSuite(t *testing.T) {
	suite.Run(t, new(LoggingTestSuite))
}

func (s *LoggingTestSuite) TestSetupLogging_Success() {
	// Arrange
	s.mockClient.On("AddPlugin", mock.Anything).Return().Once()

	// Act
	err := SetupLogging(s.mockClient, testLogFilePath)

	// Assert
	assert.NoError(s.T(), err)
	_, statErr := os.Stat(testLogFilePath)
	assert.NoError(s.T(), statErr, "Log file should be created")
}

func (s *LoggingTestSuite) TestSetupLogging_OpenFileError() {
	// Arrange
	// Intentionally using an invalid path to cause os.OpenFile to fail
	invalidLogFilePath := "/invalid_path/test_log.txt"

	// Act
	err := SetupLogging(s.mockClient, invalidLogFilePath)

	// Assert
	assert.Error(s.T(), err)
	s.mockClient.AssertNotCalled(s.T(), "AddPlugin", mock.Anything)
}

func (s *LoggingTestSuite) TestSetupLogging_FileAlreadyExists() {
	// Arrange
	// Create the file first
	file, err := os.Create(testLogFilePath)
	assert.NoError(s.T(), err)
	err = file.Close()
	assert.NoError(s.T(), err)

	s.mockClient.On("AddPlugin", mock.Anything).Return().Once()

	// Act
	err = SetupLogging(s.mockClient, testLogFilePath)

	// Assert
	assert.NoError(s.T(), err)
	_, statErr := os.Stat(testLogFilePath)
	assert.NoError(s.T(), statErr, "Log file should exist")
}
