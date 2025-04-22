package plex

type Entry struct {
	ID                  string  `json:"ratingKey"`
	GUID                string  `json:"guid"`
	Title               string  `json:"title"`
	OriginalTitle       string  `json:"originalTitle"`
	Type                string  `json:"type"`
	AudienceRatingImage string  `json:"audienceRatingImage"`
	Rating              float32 `json:"rating"`
	AudienceRating      float32 `json:"audienceRating"`
	UserRating          float32 `json:"userRating"`
	Year                int     `json:"year"`
	AddedAt             int     `json:"addedAt"`
	UpdatedAt           int     `json:"updatedAt"`
	Poster              string  `json:"thumb"`
	Media               []Media `json:"Media"`
}
