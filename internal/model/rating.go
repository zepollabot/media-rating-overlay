package model

const (
	RatingServiceTypeCritic   = "critic"
	RatingServiceTypeAudience = "audience"
	RatingServiceTypeUser     = "user"
)

type Rating struct {
	Name   string
	Rating float32
	Type   string
}
