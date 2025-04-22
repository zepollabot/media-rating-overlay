package model

import (
	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
)

const (
	MediaServicePlex     = "plex"
	MediaServiceKodi     = "kodi"
	MediaServiceJellyfin = "jellyfin"
)

var MediaServices = []string{
	MediaServicePlex,
	MediaServiceKodi,
	MediaServiceJellyfin,
}

type MediaService struct {
	Name           string
	Libraries      []config.Library
	Client         media.MediaClient
	LibraryService media.LibraryService
	ItemService    media.ItemService
	PosterService  media.PosterService
}
