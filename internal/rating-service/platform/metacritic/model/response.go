package metacritic

type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	TotalResults int     `json:"totalResults"`
	ID           string  `json:"id"`
	Items        []Entry `json:"items"`
}
