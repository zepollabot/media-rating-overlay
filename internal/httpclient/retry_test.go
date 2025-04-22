package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RetryClientTestSuite struct {
	suite.Suite
}

func TestRetryClientTestSuite(t *testing.T) {
	suite.Run(t, new(RetryClientTestSuite))
}

func (s *RetryClientTestSuite) TestNewRetryClient_ReturnsClient() {
	// Arrange
	testTimeout := 5 * time.Second
	testMaxRetries := 3

	// Act
	client := NewRetryClient(testTimeout, testMaxRetries)

	// Assert
	assert.NotNil(s.T(), client, "Client should not be nil")
}

func (s *RetryClientTestSuite) TestNewRetryClient_DifferentParameters() {
	// Arrange
	anotherTimeout := 10 * time.Second
	anotherMaxRetries := 5

	// Act
	client := NewRetryClient(anotherTimeout, anotherMaxRetries)

	// Assert
	assert.NotNil(s.T(), client, "Client should not be nil with different parameters")
}
