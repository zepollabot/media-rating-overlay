package file

import (
	"os"
	"path"
	"strings"

	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
)

// FileManager implements the FileManager interface
type FileManager struct {
	logger *zap.Logger
}

// NewFileManager creates a new file manager
func NewFileManager(logger *zap.Logger) *FileManager {
	return &FileManager{
		logger: logger,
	}
}

func (m *FileManager) CheckIfPosterExists(filePath string) (bool, error) {
	m.logger.Debug("Checking if poster exists..",
		zap.String("filePath", filePath),
	)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		m.logger.Error("unable to check if poster exists",
			zap.String("filePath", filePath),
			zap.Error(err),
		)
		return false, err
	}
	m.logger.Debug("Original poster found",
		zap.String("filePath", filePath),
	)
	return true, nil
}

func (m *FileManager) SavePoster(filePath string, data []byte) error {
	m.logger.Debug("Saving poster..",
		zap.String("filePath", filePath),
	)
	err := os.WriteFile(filePath, data, 0664)
	if err != nil {
		m.logger.Error("unable to save file",
			zap.String("filePath", filePath),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// GeneratePosterFilePath generates a poster file path from the original file path
func (m *FileManager) GeneratePosterFilePath(filePath string, ext string) string {
	m.logger.Debug("Generating poster file path..",
		zap.String("filePath", filePath),
		zap.String("ext", ext),
	)
	// change file ext to PNG
	actualExt := path.Ext(filePath)
	newFilePath := filePath[0:len(filePath)-len(actualExt)] + ext
	newFilePath = strings.Replace(newFilePath, "-original.", "-poster.", 1)

	return newFilePath
}

// BackupExistingPoster backs up an existing poster file
func (m *FileManager) BackupExistingPoster(filePath string) error {
	m.logger.Debug("Backing up existing poster..",
		zap.String("filePath", filePath),
	)
	var maybePosterPaths []string
	maybePosterPaths = append(maybePosterPaths, m.GeneratePosterFilePath(filePath, ".jpeg"))
	maybePosterPaths = append(maybePosterPaths, m.GeneratePosterFilePath(filePath, ".jpg"))

	for _, testPath := range maybePosterPaths {
		if _, errOpenMaybePoster := os.Stat(testPath); errOpenMaybePoster == nil {
			backupPosterFilePath := testPath + "-backup"
			m.logger.Debug("Existent poster found, renaming..",
				zap.String("backupPosterFilePath", backupPosterFilePath),
			)
			e := os.Rename(testPath, backupPosterFilePath)
			if e != nil {
				m.logger.Debug("unable to rename existent poster",
					zap.String("posterPath", testPath),
					zap.Error(e),
				)
				return &model.PosterError{
					Stage: "backup_poster",
					Err:   e,
				}
			}
		}
	}

	return nil
}
