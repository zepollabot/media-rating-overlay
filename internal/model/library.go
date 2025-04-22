package model

type Library struct {
	ID       string `json:"ID"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Language string `json:"language"`
}
