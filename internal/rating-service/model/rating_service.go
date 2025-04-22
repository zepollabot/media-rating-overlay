package model

import "github.com/zepollabot/media-rating-overlay/internal/rating-service"

type RatingService struct {
	Name            string
	PlatformService rating.RatingPlatformService
	LogoService     rating.LogoService
}
