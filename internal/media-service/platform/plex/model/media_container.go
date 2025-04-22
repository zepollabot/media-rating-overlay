package plex

type MediaContainer struct {
	ID        int       `json:"librarySectionID"`
	Title     string    `json:"librarySectionTitle"`
	Size      int       `json:"size"`
	Entries   []Entry   `json:"Metadata"`
	Libraries []Library `json:"Directory"`
}
