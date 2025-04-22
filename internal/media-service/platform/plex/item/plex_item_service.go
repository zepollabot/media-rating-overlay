package plex

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/constant"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	plex "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/model"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type ItemService interface {
	GetItems(ctx context.Context, library model.Library, config *config.Library) ([]model.Item, error)
}

// PlexItemService handles Plex item operations
type PlexItemService struct {
	client         media.MediaClient
	logger         *zap.Logger
	filtersService media.FilterService
}

// NewPlexItemService creates a new Plex item service
func NewPlexItemService(client media.MediaClient, logger *zap.Logger, filtersService media.FilterService) ItemService {
	return &PlexItemService{
		client:         client,
		logger:         logger,
		filtersService: filtersService,
	}
}

// GetItems retrieves items from a specific library
func (s *PlexItemService) GetItems(ctx context.Context, library model.Library, config *config.Library) ([]model.Item, error) {
	baseUrl := s.client.GetBaseUrl()
	endpoint := baseUrl.JoinPath(fmt.Sprintf("/library/sections/%s/all", library.ID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		s.logger.Error("unable to build request",
			zap.String("method", "GetEntries"),
			zap.String("url", endpoint.String()),
			zap.Error(err),
		)
		return nil, err
	}

	filters := s.filtersService.ConvertConfigFiltersToRequestFilters(config)
	s.filtersService.ApplyFiltersToRequest(req, filters)

	response, err := s.client.DoWithMediaResponse(req)
	if err != nil {
		s.logger.Error("unable to perform request to Plex Client",
			zap.String("method", "GetEntries"),
			zap.Error(err),
		)
		return nil, err
	}

	plexResponse, ok := response.(*plex.Response)
	if !ok {
		s.logger.Error("unable to cast response to plex.Response",
			zap.String("method", "GetLibraryItems"),
		)
		return nil, fmt.Errorf("invalid response type")
	}

	return s.convertPlexItems(plexResponse.MediaContainer.Entries), nil
}

// convertPlexItems converts Plex items response to common Item model
func (s *PlexItemService) convertPlexItems(items []plex.Entry) []model.Item {
	var convertedItems []model.Item
	lo.ForEach(items, func(entry plex.Entry, index int) {

		convertedItems = append(convertedItems, model.Item{
			ID:         entry.ID,
			GUID:       entry.GUID,
			Title:      entry.Title,
			Type:       entry.Type,
			Year:       entry.Year,
			Ratings:    s.buildRatings(entry),
			AddedAt:    media.ConvertoTimestampToUTC(entry.AddedAt),
			UpdatedAt:  media.ConvertoTimestampToUTC(entry.UpdatedAt),
			Poster:     entry.Poster,
			Media:      s.convertPlexMedia(entry.Media),
			IsEligible: s.isEligibleForPoster(entry.Type, entry.GUID),
		})
	})
	return convertedItems
}

// TODO	 Move outside of media service, maybe during the creation or processing of the item
func (s *PlexItemService) buildRatings(entry plex.Entry) []model.Rating {

	ratingServices := []model.Rating{}

	s.logger.Debug("Building rating services for entry", zap.Any("title", entry.Title))

	ratingServiceName := s.guessRatingService(entry.AudienceRatingImage)

	switch ratingServiceName {
	case constant.RatingServiceRottenTomatoes:
		ratingServices = append(ratingServices, model.Rating{
			Name:   constant.RatingServiceRottenTomatoes,
			Type:   model.RatingServiceTypeCritic,
			Rating: entry.Rating,
		})

		ratingServices = append(ratingServices, model.Rating{
			Name:   constant.RatingServiceRottenTomatoes,
			Type:   model.RatingServiceTypeAudience,
			Rating: entry.AudienceRating,
		})

	case constant.RatingServiceIMDB:
		ratingServices = append(ratingServices, model.Rating{
			Name:   constant.RatingServiceIMDB,
			Type:   model.RatingServiceTypeAudience,
			Rating: entry.AudienceRating,
		})

	case constant.RatingServiceTMDB:
		ratingServices = append(ratingServices, model.Rating{
			Name:   constant.RatingServiceTMDB,
			Type:   model.RatingServiceTypeAudience,
			Rating: entry.AudienceRating,
		})

	default:
	}

	return ratingServices
}

func (s *PlexItemService) guessRatingService(audienceRatingImage string) string {
	re := regexp.MustCompile(`([a-zA-Z]+)\:\/\/\S+`)

	ratingService := ""

	if re.MatchString(audienceRatingImage) {
		subMatchAll := re.FindStringSubmatch(audienceRatingImage)
		switch subMatchAll[1] {
		case "rottentomatoes":
			ratingService = constant.RatingServiceRottenTomatoes
		case "imdb":
			ratingService = constant.RatingServiceIMDB
		case "themoviedb":
			ratingService = constant.RatingServiceTMDB
		default:
			s.logger.Info("unable to guess rating service",
				zap.String("audienceRatingImage", audienceRatingImage),
				zap.String("ratingService", subMatchAll[1]),
			)
		}
	}

	return ratingService
}

func (s *PlexItemService) isEligibleForPoster(itemType string, itemGUID string) bool {
	return itemType == "movie" &&
		!strings.Contains(itemGUID, "local:")
}

func (s *PlexItemService) convertPlexMedia(media []plex.Media) []model.Media {
	convertedMedia := []model.Media{}

	for _, m := range media {
		convertedMedia = append(convertedMedia, model.Media{
			File: lo.Map(m.File, func(file plex.File, _ int) model.File {
				return model.File{
					Position: file.Position,
				}
			}),
		})
	}

	return convertedMedia
}
