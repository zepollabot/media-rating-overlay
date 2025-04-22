package tmdb

type Response struct {
	Page    int     `json:"page"`
	Results []Entry `json:"results"`
}
