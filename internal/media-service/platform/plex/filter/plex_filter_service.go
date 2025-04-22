package plex

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type Clock interface {
	Now() time.Time
}

type PlexFiltersService struct {
	clock Clock
}

func NewPlexFiltersService(clock Clock) media.FilterService {
	return &PlexFiltersService{
		clock: clock,
	}
}

func (s *PlexFiltersService) ApplyFiltersToRequest(request *http.Request, filters []model.Filter) {

	q := request.URL.Query()

	for _, filter := range filters {
		if filter.Name != "" && filter.Value != "" {
			q.Add(filter.Name, filter.Value)
		}
	}

	request.URL.RawQuery = q.Encode()
}

func (s *PlexFiltersService) buildAddedAtValue(addedAt string) string {
	re := regexp.MustCompile(`last_(\d+)_(days|months|years)`)

	if re.MatchString(addedAt) {
		subMatchAll := re.FindStringSubmatch(addedAt)

		value, _ := strconv.Atoi(subMatchAll[1])
		period := subMatchAll[2]

		var after time.Time

		now := s.clock.Now()

		switch period {
		case "days":
			after = now.AddDate(0, 0, -value)
		case "months":
			after = now.AddDate(0, -value, 0)
		case "years":
			after = now.AddDate(-value, 0, 0)
		default:
		}

		sec := after.Unix()
		return strconv.Itoa(int(sec))
	}

	return ""
}

func (s *PlexFiltersService) ConvertConfigFiltersToRequestFilters(config *config.Library) []model.Filter {
	filters := []model.Filter{}

	if len(config.Filters.Year) > 0 {
		filters = append(filters, model.Filter{
			Name:  "year",
			Value: strings.Join(config.Filters.Year, ","),
		})
	}

	if len(config.Filters.Title) > 0 {
		filters = append(filters, model.Filter{
			Name:  "title",
			Value: strings.Join(config.Filters.Title, ","),
		})
	}

	if len(config.Filters.Genre) > 0 {
		filters = append(filters, model.Filter{
			Name:  "genre",
			Value: strings.Join(config.Filters.Genre, ","),
		})
	}

	if config.Filters.AddedAt != "" {
		filters = append(filters, model.Filter{
			Name:  "addedAt>",
			Value: s.buildAddedAtValue(config.Filters.AddedAt),
		})
	}

	return filters
}
