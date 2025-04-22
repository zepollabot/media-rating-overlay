package image

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type FileManager interface {
	GeneratePosterFilePath(filePath string, ext string) string
	BackupExistingPoster(filePath string) error
}

// DefaultImageService implements the ImageService interface
type ImageService struct {
	logger      *zap.Logger
	config      *model.PosterConfig
	fileManager FileManager
}

// NewImageService creates a new image service
func NewImageService(logger *zap.Logger, config *model.PosterConfig, fileManager FileManager) *ImageService {
	return &ImageService{
		logger:      logger,
		config:      config,
		fileManager: fileManager,
	}
}

// OpenImage opens an image file
func (p *ImageService) OpenImage(filePath string) (image.Image, error) {
	img, err := imaging.Open(filePath)
	if err != nil {
		p.logger.Error(
			"failed to open image",
			zap.String("filePath", filePath),
			zap.Error(err),
		)
		return nil, &model.PosterError{
			Stage: "open_image",
			Err:   err,
		}
	}
	return img, nil
}

// ResizeImage resizes an image
func (p *ImageService) ResizeImage(img image.Image, width int, height int) image.Image {
	return imaging.Resize(img, width, height, imaging.Lanczos)
}

// SaveImage saves an image to a file
func (p *ImageService) SaveImage(img image.Image, filePath string) (string, error) {

	// Generate poster file path
	posterFilePath := p.fileManager.GeneratePosterFilePath(filePath, ".png")

	// Backup existing poster
	if err := p.fileManager.BackupExistingPoster(posterFilePath); err != nil {
		p.logger.Warn("Failed to backup existing poster",
			zap.String("filePath", filePath),
			zap.Error(err),
		)
	}

	p.logger.Debug("Writing new poster..",
		zap.String("filePath", posterFilePath),
	)

	if err := p.SaveImageObject(img, posterFilePath); err != nil {
		return "", err
	}

	return posterFilePath, nil
}

// CreateContext creates a new drawing context
func (p *ImageService) CreateContext(width, height int) *gg.Context {
	dc := gg.NewContext(width, height)

	// Set a black background
	dc.SetColor(color.RGBA{A: 0xFF})
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Fill()

	return dc
}

func (p *ImageService) SaveImageObject(img image.Image, filePath string) error {
	if err := imaging.Save(img, filePath); err != nil {
		p.logger.Error(
			"unable to save image",
			zap.String("filePath", filePath),
			zap.Error(err),
		)
		return err
	}

	// Set correct permissions for the saved image
	if err := os.Chmod(filePath, 0644); err != nil {
		p.logger.Error(
			"unable to set image permissions",
			zap.String("filePath", filePath),
			zap.Error(err),
		)
		return err
	}

	return nil
}
