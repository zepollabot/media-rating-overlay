package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/processor/poster"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

type PosterGeneratorFactoryTestSuite struct {
	suite.Suite
	factory *PosterGeneratorFactory
	logger  *zap.Logger
}

func (s *PosterGeneratorFactoryTestSuite) SetupSuite() {
	s.logger = zap.NewNop()
	s.factory = NewPosterGeneratorFactory(s.logger, []ratingModel.RatingService{}, false)
}

func (s *PosterGeneratorFactoryTestSuite) TestCreate() {
	// Act
	generator := s.factory.Create()

	// Assert
	s.Require().NotNil(generator)
	s.IsType(&poster.PosterGenerator{}, generator)
}

func (s *PosterGeneratorFactoryTestSuite) TestCreateWithVisualDebug() {
	// Arrange
	factory := NewPosterGeneratorFactory(s.logger, []ratingModel.RatingService{}, true)

	// Act
	generator := factory.Create()

	// Assert
	s.Require().NotNil(generator)
	s.IsType(&poster.PosterGenerator{}, generator)
}

func TestPosterGeneratorFactorySuite(t *testing.T) {
	suite.Run(t, new(PosterGeneratorFactoryTestSuite))
}
