package plex_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	configmodel "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/constant"
	mediamocks "github.com/zepollabot/media-rating-overlay/internal/media-service/mocks"
	plexitem "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/item"
	plexmodel "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/model"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type PlexItemServiceTestSuite struct {
	suite.Suite
	mockMediaClient   *mediamocks.MediaClient
	mockFilterService *mediamocks.FilterService
	logger            *zap.Logger
	service           plexitem.ItemService
}

func (s *PlexItemServiceTestSuite) SetupTest() {
	s.mockMediaClient = mediamocks.NewMediaClient(s.T())
	s.mockFilterService = mediamocks.NewFilterService(s.T())
	s.logger = zap.NewNop()
	s.service = plexitem.NewPlexItemService(s.mockMediaClient, s.logger, s.mockFilterService)
}

func (s *PlexItemServiceTestSuite) TearDownTest() {
	s.mockMediaClient.AssertExpectations(s.T())
	s.mockFilterService.AssertExpectations(s.T())
}

func TestPlexItemServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PlexItemServiceTestSuite))
}

func (s *PlexItemServiceTestSuite) TestGetItems_Success() {
	// Arrange
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{
		Filters: configmodel.Filter{Genre: []string{"Action"}},
	}

	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := "http://localhost:32400/library/sections/1/all"

	mockRequestFilters := []model.Filter{{Name: "genre", Value: "Action"}}
	plexEntries := []plexmodel.Entry{
		{
			ID:                  "101",
			GUID:                "com.plexapp.agents.imdb://tt0120737?lang=en",
			Title:               "The Lord of the Rings: The Fellowship of the Ring",
			Type:                "movie",
			Year:                2001,
			AudienceRating:      9.2,
			AudienceRatingImage: "imdb://image.png",
			Rating:              8.8,
			AddedAt:             1609459200, // 2021-01-01 00:00:00 UTC
			UpdatedAt:           1609545600, // 2021-01-02 00:00:00 UTC
			Poster:              "/library/metadata/101/thumb/1609545600",
			Media: []plexmodel.Media{
				{File: []plexmodel.File{{Position: "1"}, {Position: "2"}}},
			},
		},
		{
			ID:                  "102",
			GUID:                "com.plexapp.agents.themoviedb://121?lang=en",
			Title:               "The Lord of the Rings: The Two Towers",
			Type:                "movie",
			Year:                2002,
			AudienceRating:      9.5,
			AudienceRatingImage: "themoviedb://image.png",
			Rating:              0, // No critic rating
			AddedAt:             1609632000,
			UpdatedAt:           1609718400,
			Poster:              "/library/metadata/102/thumb/1609718400",
		},
		{
			ID:                  "103",
			GUID:                "com.plexapp.agents.rottentomatoes://m/12345?lang=en",
			Title:               "Inception",
			Type:                "movie",
			Year:                2010,
			AudienceRating:      9.1,
			AudienceRatingImage: "rottentomatoes://image.png",
			Rating:              8.7, // Critic rating
			AddedAt:             1609804800,
			UpdatedAt:           1609891200,
			Poster:              "/library/metadata/103/thumb/1609891200",
		},
		{
			ID:        "104",
			GUID:      "local://something", // local item, not eligible for poster
			Title:     "Home Video",
			Type:      "movie",
			Year:      2020,
			AddedAt:   1609977600,
			UpdatedAt: 1610064000,
			Poster:    "/library/metadata/104/thumb/1610064000",
		},
	}
	plexResponse := &plexmodel.Response{MediaContainer: plexmodel.MediaContainer{Entries: plexEntries}}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return(mockRequestFilters)
	s.mockFilterService.On("ApplyFiltersToRequest", mock.AnythingOfType("*http.Request"), mockRequestFilters).Run(func(args mock.Arguments) {
		req := args.Get(0).(*http.Request)
		s.Equal(expectedEndpoint, req.URL.String())
	})
	s.mockMediaClient.On("DoWithMediaResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.String() == expectedEndpoint
	})).Return(plexResponse, nil)

	// Act
	items, err := s.service.GetItems(ctx, library, libConfig)

	// Assert
	s.NoError(err)
	s.NotNil(items)
	s.Len(items, 4)

	// Item 1: IMDB
	s.Equal("101", items[0].ID)
	s.Equal("com.plexapp.agents.imdb://tt0120737?lang=en", items[0].GUID)
	s.Equal("The Lord of the Rings: The Fellowship of the Ring", items[0].Title)
	s.Equal("movie", items[0].Type)
	s.Equal(2001, items[0].Year)
	s.True(items[0].IsEligible)
	s.Len(items[0].Ratings, 1)
	s.Equal(constant.RatingServiceIMDB, items[0].Ratings[0].Name)
	s.Equal(model.RatingServiceTypeAudience, items[0].Ratings[0].Type)
	s.Equal(float32(9.2), items[0].Ratings[0].Rating)
	s.Equal(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), items[0].AddedAt)
	s.Equal(time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC), items[0].UpdatedAt)
	s.Equal("/library/metadata/101/thumb/1609545600", items[0].Poster)
	s.Len(items[0].Media, 1)
	s.Len(items[0].Media[0].File, 2)
	s.Equal("1", items[0].Media[0].File[0].Position)

	// Item 2: TMDB
	s.Equal("102", items[1].ID)
	s.Equal("com.plexapp.agents.themoviedb://121?lang=en", items[1].GUID)
	s.True(items[1].IsEligible)
	s.Len(items[1].Ratings, 1)
	s.Equal(constant.RatingServiceTMDB, items[1].Ratings[0].Name)
	s.Equal(model.RatingServiceTypeAudience, items[1].Ratings[0].Type)
	s.Equal(float32(9.5), items[1].Ratings[0].Rating)

	// Item 3: RottenTomatoes
	s.Equal("103", items[2].ID)
	s.Equal("com.plexapp.agents.rottentomatoes://m/12345?lang=en", items[2].GUID)
	s.True(items[2].IsEligible)
	s.Len(items[2].Ratings, 2)
	foundRTCritic := false
	foundRTAudience := false
	for _, r := range items[2].Ratings {
		if r.Name == constant.RatingServiceRottenTomatoes && r.Type == model.RatingServiceTypeCritic {
			s.Equal(float32(8.7), r.Rating)
			foundRTCritic = true
		}
		if r.Name == constant.RatingServiceRottenTomatoes && r.Type == model.RatingServiceTypeAudience {
			s.Equal(float32(9.1), r.Rating)
			foundRTAudience = true
		}
	}
	s.True(foundRTCritic, "Rotten Tomatoes critic rating not found")
	s.True(foundRTAudience, "Rotten Tomatoes audience rating not found")

	// Item 4: Local, not eligible
	s.Equal("104", items[3].ID)
	s.Equal("local://something", items[3].GUID)
	s.False(items[3].IsEligible)
	s.Empty(items[3].Ratings) // No AudienceRatingImage, so no ratings parsed
}

func (s *PlexItemServiceTestSuite) TestGetItems_Error_NewRequest() {
	// Arrange
	ctx := context.Background()
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}
	baseURL, _ := url.Parse("http://localhost:32400")

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return([]model.Filter{})
	s.mockFilterService.On("ApplyFiltersToRequest", mock.Anything, []model.Filter{})
	s.mockMediaClient.On("DoWithMediaResponse", mock.Anything).Return(nil, errors.New("context canceled"))

	// Act
	items, err := s.service.GetItems(cancelledCtx, library, libConfig)

	// Assert
	s.Error(err)
	s.Nil(items)
	s.Contains(err.Error(), "context canceled")
}

func (s *PlexItemServiceTestSuite) TestGetItems_Error_ClientDo() {
	// Arrange
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}

	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := "http://localhost:32400/library/sections/1/all"
	expectedError := errors.New("client do error")

	mockRequestFilters := []model.Filter{}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return(mockRequestFilters)
	s.mockFilterService.On("ApplyFiltersToRequest", mock.AnythingOfType("*http.Request"), mockRequestFilters)
	s.mockMediaClient.On("DoWithMediaResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.String() == expectedEndpoint
	})).Return(nil, expectedError)

	// Act
	items, err := s.service.GetItems(ctx, library, libConfig)

	// Assert
	s.Error(err)
	s.Equal(expectedError, err)
	s.Nil(items)
}

func (s *PlexItemServiceTestSuite) TestGetItems_Error_InvalidResponseType() {
	// Arrange
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}

	baseURL, _ := url.Parse("http://localhost:32400")
	expectedEndpoint := "http://localhost:32400/library/sections/1/all"

	invalidResponse := &struct{}{}

	mockRequestFilters := []model.Filter{}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return(mockRequestFilters)
	s.mockFilterService.On("ApplyFiltersToRequest", mock.AnythingOfType("*http.Request"), mockRequestFilters)
	s.mockMediaClient.On("DoWithMediaResponse", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.String() == expectedEndpoint
	})).Return(invalidResponse, nil)

	// Act
	items, err := s.service.GetItems(ctx, library, libConfig)

	// Assert
	s.Error(err)
	s.Nil(items)
	s.EqualError(err, "invalid response type")
}

func (s *PlexItemServiceTestSuite) TestGetItems_GuessRatingServiceEdgeCases() {
	// Arrange
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}
	baseURL, _ := url.Parse("http://localhost:32400")

	plexEntries := []plexmodel.Entry{
		{
			ID: "201", Title: "Unknown Service", Type: "movie", GUID: "com.plexapp.agents.custom://tt1",
			AudienceRatingImage: "unknown://image.png", AudienceRating: 5.0,
		},
		{
			ID: "202", Title: "Malformed Image", Type: "movie", GUID: "com.plexapp.agents.custom://tt2",
			AudienceRatingImage: "not a url", AudienceRating: 6.0,
		},
		{
			ID: "203", Title: "Empty Image", Type: "movie", GUID: "com.plexapp.agents.custom://tt3",
			AudienceRatingImage: "", AudienceRating: 7.0,
		},
	}
	plexResponse := &plexmodel.Response{MediaContainer: plexmodel.MediaContainer{Entries: plexEntries}}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return([]model.Filter{})
	s.mockFilterService.On("ApplyFiltersToRequest", mock.Anything, []model.Filter{})
	s.mockMediaClient.On("DoWithMediaResponse", mock.Anything).Return(plexResponse, nil)

	// Act
	items, err := s.service.GetItems(ctx, library, libConfig)

	// Assert
	s.NoError(err)
	s.Len(items, 3)
	s.Empty(items[0].Ratings, "Ratings should be empty for unknown service")
	s.Empty(items[1].Ratings, "Ratings should be empty for malformed AudienceRatingImage")
	s.Empty(items[2].Ratings, "Ratings should be empty for empty AudienceRatingImage")
}

func (s *PlexItemServiceTestSuite) TestGetItems_IsEligibleForPosterEdgeCases() {
	// Arrange
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}
	baseURL, _ := url.Parse("http://localhost:32400")

	plexEntries := []plexmodel.Entry{
		{ID: "301", Title: "Eligible Movie", Type: "movie", GUID: "com.plexapp.agents.imdb://tt1"},
		{ID: "302", Title: "Ineligible TV Show", Type: "show", GUID: "com.plexapp.agents.imdb://tt2"},
		{ID: "303", Title: "Ineligible Local Movie", Type: "movie", GUID: "local://movie"},
	}
	plexResponse := &plexmodel.Response{MediaContainer: plexmodel.MediaContainer{Entries: plexEntries}}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return([]model.Filter{})
	s.mockFilterService.On("ApplyFiltersToRequest", mock.Anything, []model.Filter{})
	s.mockMediaClient.On("DoWithMediaResponse", mock.Anything).Return(plexResponse, nil)

	// Act
	items, err := s.service.GetItems(ctx, library, libConfig)

	// Assert
	s.NoError(err)
	s.Len(items, 3)
	s.True(items[0].IsEligible, "Movie with external GUID should be eligible")
	s.False(items[1].IsEligible, "TV show should not be eligible")
	s.False(items[2].IsEligible, "Movie with local GUID should not be eligible")
}

func (s *PlexItemServiceTestSuite) TestGetItems_EmptyPlexResponse() {
	// Arrange
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}
	baseURL, _ := url.Parse("http://localhost:32400")

	plexResponse := &plexmodel.Response{MediaContainer: plexmodel.MediaContainer{Entries: []plexmodel.Entry{}}}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return([]model.Filter{})
	s.mockFilterService.On("ApplyFiltersToRequest", mock.Anything, []model.Filter{})
	s.mockMediaClient.On("DoWithMediaResponse", mock.Anything).Return(plexResponse, nil)

	// Act
	items, err := s.service.GetItems(ctx, library, libConfig)

	// Assert
	s.NoError(err)
	s.Empty(items, "Items slice should be empty when Plex returns no entries")
}

func (s *PlexItemServiceTestSuite) TestGetItems_NilPlexResponseMediaContainer() {
	// Arrange
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}
	baseURL, _ := url.Parse("http://localhost:32400")

	plexResponse := &plexmodel.Response{}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return([]model.Filter{})
	s.mockFilterService.On("ApplyFiltersToRequest", mock.Anything, []model.Filter{})
	s.mockMediaClient.On("DoWithMediaResponse", mock.Anything).Return(plexResponse, nil)

	// Act
	items, err := s.service.GetItems(ctx, library, libConfig)

	// Assert
	s.NoError(err)
	s.Empty(items, "Items slice should be empty when Plex returns a response with nil MediaContainer.Entries")
}

func (s *PlexItemServiceTestSuite) TestGetItems_RatingLogic() {
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}
	baseURL, _ := url.Parse("http://localhost:32400")

	plexEntries := []plexmodel.Entry{
		{
			ID: "rt001", Title: "RT Movie", Type: "movie", GUID: "com.plexapp.agents.rt://1",
			AudienceRatingImage: "rottentomatoes://image.png", AudienceRating: 9.0, Rating: 8.0,
		},
		{
			ID: "imdb001", Title: "IMDB Movie", Type: "movie", GUID: "com.plexapp.agents.imdb://1",
			AudienceRatingImage: "imdb://image.png", AudienceRating: 7.5, Rating: 0,
		},
		{
			ID: "tmdb001", Title: "TMDB Movie", Type: "movie", GUID: "com.plexapp.agents.tmdb://1",
			AudienceRatingImage: "themoviedb://image.png", AudienceRating: 8.5,
		},
		{
			ID: "unknown001", Title: "Unknown Rating Movie", Type: "movie", GUID: "com.plexapp.agents.other://1",
			AudienceRatingImage: "someother://image.png", AudienceRating: 6.0,
		},
		{
			ID: "rt002", Title: "RT Movie Zero Critic", Type: "movie", GUID: "com.plexapp.agents.rt://2",
			AudienceRatingImage: "rottentomatoes://image.png", AudienceRating: 7.0, Rating: 0,
		},
	}
	plexResponse := &plexmodel.Response{MediaContainer: plexmodel.MediaContainer{Entries: plexEntries}}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return([]model.Filter{})
	s.mockFilterService.On("ApplyFiltersToRequest", mock.Anything, []model.Filter{})
	s.mockMediaClient.On("DoWithMediaResponse", mock.Anything).Return(plexResponse, nil)

	items, err := s.service.GetItems(ctx, library, libConfig)
	s.NoError(err)
	s.Len(items, 5)

	// RT Movie
	s.Len(items[0].Ratings, 2)
	foundRTCritic, foundRTAudience := false, false
	for _, r := range items[0].Ratings {
		if r.Name == constant.RatingServiceRottenTomatoes && r.Type == model.RatingServiceTypeCritic && r.Rating == 8.0 {
			foundRTCritic = true
		}
		if r.Name == constant.RatingServiceRottenTomatoes && r.Type == model.RatingServiceTypeAudience && r.Rating == 9.0 {
			foundRTAudience = true
		}
	}
	s.True(foundRTCritic, "RT Critic rating missing or incorrect for rt001")
	s.True(foundRTAudience, "RT Audience rating missing or incorrect for rt001")

	// IMDB Movie
	s.Len(items[1].Ratings, 1)
	s.Equal(constant.RatingServiceIMDB, items[1].Ratings[0].Name)
	s.Equal(model.RatingServiceTypeAudience, items[1].Ratings[0].Type)
	s.Equal(float32(7.5), items[1].Ratings[0].Rating)

	// TMDB Movie
	s.Len(items[2].Ratings, 1)
	s.Equal(constant.RatingServiceTMDB, items[2].Ratings[0].Name)
	s.Equal(model.RatingServiceTypeAudience, items[2].Ratings[0].Type)
	s.Equal(float32(8.5), items[2].Ratings[0].Rating)

	// Unknown Rating Movie
	s.Empty(items[3].Ratings)

	// RT Movie Zero Critic
	s.Len(items[4].Ratings, 2)
	foundRTCriticZero, foundRTAudienceZero := false, false
	for _, r := range items[4].Ratings {
		if r.Name == constant.RatingServiceRottenTomatoes && r.Type == model.RatingServiceTypeCritic && r.Rating == 0.0 {
			foundRTCriticZero = true
		}
		if r.Name == constant.RatingServiceRottenTomatoes && r.Type == model.RatingServiceTypeAudience && r.Rating == 7.0 {
			foundRTAudienceZero = true
		}
	}
	s.True(foundRTCriticZero, "RT Critic rating (zero) missing or incorrect for rt002")
	s.True(foundRTAudienceZero, "RT Audience rating missing or incorrect for rt002")
}

// Test Media Conversion specifically
func (s *PlexItemServiceTestSuite) TestGetItems_MediaConversion() {
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}
	baseURL, _ := url.Parse("http://localhost:32400")

	plexEntries := []plexmodel.Entry{
		{
			ID: "media001", Title: "Media Test", Type: "movie", GUID: "com.plexapp.agents.imdb://tt1",
			Media: []plexmodel.Media{
				{File: []plexmodel.File{{Position: "1"}, {Position: "3"}}},
				{File: []plexmodel.File{{Position: "0"}}},
			},
		},
		{
			ID: "media002", Title: "No Media Test", Type: "movie", GUID: "com.plexapp.agents.imdb://tt2",
			Media: []plexmodel.Media{},
		},
		{
			ID: "media003", Title: "Nil Files Test", Type: "movie", GUID: "com.plexapp.agents.imdb://tt3",
			Media: []plexmodel.Media{{File: nil}},
		},
	}
	plexResponse := &plexmodel.Response{MediaContainer: plexmodel.MediaContainer{Entries: plexEntries}}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return([]model.Filter{})
	s.mockFilterService.On("ApplyFiltersToRequest", mock.Anything, []model.Filter{})
	s.mockMediaClient.On("DoWithMediaResponse", mock.Anything).Return(plexResponse, nil)

	items, err := s.service.GetItems(ctx, library, libConfig)
	s.NoError(err)
	s.Len(items, 3)

	// media001
	s.Len(items[0].Media, 2)
	s.Len(items[0].Media[0].File, 2)
	s.Equal("1", items[0].Media[0].File[0].Position)
	s.Equal("3", items[0].Media[0].File[1].Position)
	s.Len(items[0].Media[1].File, 1)
	s.Equal("0", items[0].Media[1].File[0].Position)

	// media002
	s.Empty(items[1].Media)

	// media003
	s.Len(items[2].Media, 1)
	s.Empty(items[2].Media[0].File)
}

// Test for timestamps conversion
func (s *PlexItemServiceTestSuite) TestGetItems_TimestampConversion() {
	ctx := context.Background()
	library := model.Library{ID: "1"}
	libConfig := &configmodel.Library{}
	baseURL, _ := url.Parse("http://localhost:32400")

	plexEntries := []plexmodel.Entry{
		{
			ID: "time001", Title: "Timestamp Test", Type: "movie", GUID: "com.plexapp.agents.imdb://tt1",
			AddedAt:   1678881600, // 2023-03-15 12:00:00 UTC
			UpdatedAt: 0,
		},
	}
	plexResponse := &plexmodel.Response{MediaContainer: plexmodel.MediaContainer{Entries: plexEntries}}

	s.mockMediaClient.On("GetBaseUrl").Return(baseURL)
	s.mockFilterService.On("ConvertConfigFiltersToRequestFilters", libConfig).Return([]model.Filter{})
	s.mockFilterService.On("ApplyFiltersToRequest", mock.Anything, []model.Filter{})
	s.mockMediaClient.On("DoWithMediaResponse", mock.Anything).Return(plexResponse, nil)

	items, err := s.service.GetItems(ctx, library, libConfig)
	s.NoError(err)
	s.Len(items, 1)

	expectedAddedAt := time.Date(2023, 3, 15, 12, 0, 0, 0, time.UTC)
	s.Equal(expectedAddedAt, items[0].AddedAt)
	s.True(items[0].UpdatedAt.IsZero(), "UpdatedAt should be zero time if Plex sends 0")
}
