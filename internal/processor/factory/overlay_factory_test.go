package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	overlay "github.com/zepollabot/media-rating-overlay/internal/processor/overlay"
)

type OverlayFactoryTestSuite struct {
	suite.Suite
	factory *OverlayFactory
	logger  *zap.Logger
	config  *model.PosterConfig
}

func (s *OverlayFactoryTestSuite) SetupSuite() {
	s.logger = zap.NewNop()
	s.config = model.PosterConfigWithDefaultValues()
	s.factory = NewOverlayFactory(s.logger, s.config)
}

func (s *OverlayFactoryTestSuite) TestCreateFrameOverlay() {
	// Act
	overlayFrame, err := s.factory.CreateOverlay("frame")

	// Assert
	s.Require().NoError(err)
	s.Require().NotNil(overlayFrame)
	s.IsType(&overlay.FrameOverlay{}, overlayFrame)
}

func (s *OverlayFactoryTestSuite) TestCreateBarOverlay() {
	// Act
	overlayBar, err := s.factory.CreateOverlay("bar")

	// Assert
	s.Require().NoError(err)
	s.Require().NotNil(overlayBar)
	s.IsType(&overlay.BarOverlay{}, overlayBar)
}

func TestOverlayFactorySuite(t *testing.T) {
	suite.Run(t, new(OverlayFactoryTestSuite))
}
