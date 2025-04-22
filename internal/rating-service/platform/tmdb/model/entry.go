package tmdb

type Entry struct {
	ID            int     `json:"id"`
	Adult         bool    `json:"adult"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Vote          float64 `json:"vote_average"`
}
