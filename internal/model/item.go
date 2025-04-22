package model

import (
	"time"
)

type Item struct {
	ID         string
	GUID       string
	Title      string
	Type       string
	Year       int
	Ratings    []Rating
	AddedAt    time.Time
	UpdatedAt  time.Time
	Poster     string
	Media      []Media
	IsEligible bool
}
