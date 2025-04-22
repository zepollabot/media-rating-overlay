package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (s *UtilsTestSuite) TestGetExtensionByMimeType_Success() {
	// Arrange
	mimeType := "image/jpeg"
	expectedExtension := ".jpeg"

	// Act
	actualExtension, err := GetExtensionByMimeType(mimeType)

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedExtension, actualExtension)
}

func (s *UtilsTestSuite) TestGetExtensionByMimeType_NotFound() {
	// Arrange
	mimeType := "application/json"
	expectedErrorMsg := "unable to find extension for mime type application/json"

	// Act
	actualExtension, err := GetExtensionByMimeType(mimeType)

	// Assert
	assert.Error(s.T(), err)
	assert.EqualError(s.T(), err, expectedErrorMsg)
	assert.Empty(s.T(), actualExtension)
}
