package plex

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	plex_mocks "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/filter/mocks"
	appmodel "github.com/zepollabot/media-rating-overlay/internal/model"
)

type PlexFiltersServiceTestSuite struct {
	suite.Suite
	mockClock *plex_mocks.Clock
	service   media.FilterService
}

func (s *PlexFiltersServiceTestSuite) SetupTest() {
	s.mockClock = plex_mocks.NewClock(s.T())
	s.service = NewPlexFiltersService(s.mockClock)
}

func (s *PlexFiltersServiceTestSuite) TearDownTest() {
	s.mockClock.AssertExpectations(s.T())
}

func TestPlexFiltersServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PlexFiltersServiceTestSuite))
}

func (s *PlexFiltersServiceTestSuite) TestNewPlexFiltersService() {
	// Arrange
	mockClock := plex_mocks.NewClock(s.T())

	// Act
	service := NewPlexFiltersService(mockClock)

	// Assert
	assert.NotNil(s.T(), service)
	// We can't directly assert the internal clock field as it's unexported.
	// We can indirectly test its usage in other methods or ensure the type is correct.
	_, ok := service.(*PlexFiltersService)
	assert.True(s.T(), ok, "Service should be of type *plex.PlexFiltersService")
}

func (s *PlexFiltersServiceTestSuite) TestApplyFiltersToRequest() {
	tests := []struct {
		name           string
		initialURL     string
		filters        []appmodel.Filter
		expectedURLRaw string
	}{
		{
			name:           "no filters",
			initialURL:     "http://localhost?a=b",
			filters:        []appmodel.Filter{},
			expectedURLRaw: "a=b",
		},
		{
			name:       "single filter",
			initialURL: "http://localhost",
			filters: []appmodel.Filter{
				{Name: "type", Value: "movie"},
			},
			expectedURLRaw: "type=movie",
		},
		{
			name:       "multiple filters",
			initialURL: "http://localhost?existing=true",
			filters: []appmodel.Filter{
				{Name: "year", Value: "2023"},
				{Name: "genre", Value: "action"},
			},
			expectedURLRaw: "existing=true&genre=action&year=2023", // Order might vary, url.Values.Encode() sorts by key
		},
		{
			name:       "filter with empty name",
			initialURL: "http://localhost",
			filters: []appmodel.Filter{
				{Name: "", Value: "movie"},
			},
			expectedURLRaw: "",
		},
		{
			name:       "filter with empty value",
			initialURL: "http://localhost",
			filters: []appmodel.Filter{
				{Name: "type", Value: ""},
			},
			expectedURLRaw: "",
		},
		{
			name:       "filter with empty name and value",
			initialURL: "http://localhost",
			filters: []appmodel.Filter{
				{Name: "", Value: ""},
			},
			expectedURLRaw: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Arrange
			reqURL, _ := url.Parse(tc.initialURL)
			req := &http.Request{URL: reqURL}

			// Act
			s.service.ApplyFiltersToRequest(req, tc.filters)

			// Assert
			assert.Equal(s.T(), tc.expectedURLRaw, req.URL.RawQuery)
		})
	}
}

func (s *PlexFiltersServiceTestSuite) TestConvertConfigFiltersToRequestFilters() {
	// Arrange
	now := time.Date(2024, time.July, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		configLibrary   *config.Library
		mockNow         time.Time
		expectedFilters []appmodel.Filter
	}{
		{
			name:            "empty config",
			configLibrary:   &config.Library{Filters: config.Filter{}},
			mockNow:         now, // Not strictly needed here but good for consistency
			expectedFilters: []appmodel.Filter{},
		},
		{
			name: "all filter types present",
			configLibrary: &config.Library{
				Filters: config.Filter{
					Year:    []string{"2022", "2023"},
					Title:   []string{"Movie A", "Movie B"},
					Genre:   []string{"Action", "Comedy"},
					AddedAt: "last_7_days",
				},
			},
			mockNow: now,
			expectedFilters: []appmodel.Filter{
				{Name: "year", Value: "2022,2023"},
				{Name: "title", Value: "Movie A,Movie B"},
				{Name: "genre", Value: "Action,Comedy"},
				{Name: "addedAt>", Value: strconv.FormatInt(now.AddDate(0, 0, -7).Unix(), 10)},
			},
		},
		{
			name: "only year filter",
			configLibrary: &config.Library{
				Filters: config.Filter{
					Year: []string{"2024"},
				},
			},
			mockNow: now,
			expectedFilters: []appmodel.Filter{
				{Name: "year", Value: "2024"},
			},
		},
		{
			name: "only title filter",
			configLibrary: &config.Library{
				Filters: config.Filter{
					Title: []string{"Specific Title"},
				},
			},
			mockNow: now,
			expectedFilters: []appmodel.Filter{
				{Name: "title", Value: "Specific Title"},
			},
		},
		{
			name: "only genre filter",
			configLibrary: &config.Library{
				Filters: config.Filter{
					Genre: []string{"Drama"},
				},
			},
			mockNow: now,
			expectedFilters: []appmodel.Filter{
				{Name: "genre", Value: "Drama"},
			},
		},
		{
			name: "only AddedAt filter - last 2 months",
			configLibrary: &config.Library{
				Filters: config.Filter{
					AddedAt: "last_2_months",
				},
			},
			mockNow: now,
			expectedFilters: []appmodel.Filter{
				{Name: "addedAt>", Value: strconv.FormatInt(now.AddDate(0, -2, 0).Unix(), 10)},
			},
		},
		{
			name: "only AddedAt filter - last 1 year",
			configLibrary: &config.Library{
				Filters: config.Filter{
					AddedAt: "last_1_years",
				},
			},
			mockNow: now,
			expectedFilters: []appmodel.Filter{
				{Name: "addedAt>", Value: strconv.FormatInt(now.AddDate(-1, 0, 0).Unix(), 10)},
			},
		},
		{
			name: "only AddedAt filter - last 30 days - double characters number",
			configLibrary: &config.Library{
				Filters: config.Filter{
					AddedAt: "last_30_days",
				},
			},
			mockNow: now,
			expectedFilters: []appmodel.Filter{
				{Name: "addedAt>", Value: strconv.FormatInt(now.AddDate(0, 0, -30).Unix(), 10)},
			},
		},
		{
			name: "AddedAt filter with invalid format",
			configLibrary: &config.Library{
				Filters: config.Filter{
					AddedAt: "invalid_format",
				},
			},
			mockNow: now,
			expectedFilters: []appmodel.Filter{
				{Name: "addedAt>", Value: ""}, // buildAddedAtValue returns empty for invalid
			},
		},
		{
			name: "AddedAt filter empty string",
			configLibrary: &config.Library{
				Filters: config.Filter{ // This will result in no AddedAt filter being added
					AddedAt: "",
				},
			},
			mockNow:         now,
			expectedFilters: []appmodel.Filter{ // No filter added if AddedAt is empty
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Reset mocks for each sub-test to ensure clean state
			s.mockClock.ExpectedCalls = nil // Clear previous expectations
			s.mockClock.Calls = nil         // Clear previous calls

			// Arrange
			// Only set up mock if AddedAt is present and potentially valid, to avoid unnecessary mock calls
			if tc.configLibrary.Filters.AddedAt != "" && strings.Contains(tc.configLibrary.Filters.AddedAt, "last_") {
				s.mockClock.On("Now").Return(tc.mockNow).Once()
			}

			// Act
			actualFilters := s.service.ConvertConfigFiltersToRequestFilters(tc.configLibrary)

			// Assert
			assert.ElementsMatch(s.T(), tc.expectedFilters, actualFilters)
			s.mockClock.AssertExpectations(s.T()) // Assert that all expected calls were made for this sub-test
		})
	}
}
