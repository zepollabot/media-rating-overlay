package metacritic

type Entry struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	TypeID      int    `json:"typeId"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	CriticScore struct {
		URL   string `json:"url"`
		Score int    `json:"score"`
	} `json:"criticScoreSummary"`
	Rating       string   `json:"rating"`
	ReleaseDate  string   `json:"releaseDate"`
	PremiereYear int      `json:"premiereYear"`
	Genres       []Genre  `json:"genres"`
	Platforms    []string `json:"platforms"`
	Description  string   `json:"description"`
	Duration     int      `json:"duration"`
}

type Genre struct {
	ID   *int   `json:"id"`
	Name string `json:"name"`
}
