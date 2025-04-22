package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type FileManagerTestSuite struct {
	suite.Suite
	manager *FileManager
	logger  *zap.Logger
	tempDir string
}

func (s *FileManagerTestSuite) SetupSuite() {
	s.logger = zap.NewNop()
	s.manager = NewFileManager(s.logger)
}

func (s *FileManagerTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "file-manager-test-*")
	s.Require().NoError(err)
}

func (s *FileManagerTestSuite) TearDownTest() {
	err := os.RemoveAll(s.tempDir)
	s.Require().NoError(err)
}

func (s *FileManagerTestSuite) TestCheckIfPosterExists() {
	// Arrange
	testFilePath := filepath.Join(s.tempDir, "test-poster.jpg")
	_, err := os.Create(testFilePath)
	s.Require().NoError(err)

	// Act
	exists, err := s.manager.CheckIfPosterExists(testFilePath)

	// Assert
	s.Require().NoError(err)
	s.True(exists)

	// Test non-existent file
	nonExistentPath := filepath.Join(s.tempDir, "non-existent.jpg")
	exists, err = s.manager.CheckIfPosterExists(nonExistentPath)
	s.Require().NoError(err)
	s.False(exists)
}

func (s *FileManagerTestSuite) TestSavePoster() {
	// Arrange
	testFilePath := filepath.Join(s.tempDir, "test-poster.jpg")
	testData := []byte("test poster data")

	// Act
	err := s.manager.SavePoster(testFilePath, testData)

	// Assert
	s.Require().NoError(err)
	savedData, err := os.ReadFile(testFilePath)
	s.Require().NoError(err)
	s.Equal(testData, savedData)
}

func (s *FileManagerTestSuite) TestGeneratePosterFilePath() {
	// Arrange
	originalPath := filepath.Join(s.tempDir, "test-original.jpg")

	// Act
	posterPath := s.manager.GeneratePosterFilePath(originalPath, ".png")

	// Assert
	expectedPath := filepath.Join(s.tempDir, "test-poster.png")
	s.Equal(expectedPath, posterPath)
}

func (s *FileManagerTestSuite) TestBackupExistingPoster() {
	// Arrange
	originalPath := filepath.Join(s.tempDir, "test-original.jpg")
	posterPath := s.manager.GeneratePosterFilePath(originalPath, ".jpg")
	_, err := os.Create(posterPath)
	s.Require().NoError(err)

	// Act
	err = s.manager.BackupExistingPoster(originalPath)

	// Assert
	s.Require().NoError(err)
	backupPath := posterPath + "-backup"
	_, err = os.Stat(backupPath)
	s.Require().NoError(err)
}

func TestFileManagerSuite(t *testing.T) {
	suite.Run(t, new(FileManagerTestSuite))
}
