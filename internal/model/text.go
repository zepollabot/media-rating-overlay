package model

import "github.com/fogleman/gg"

type Text struct {
	Context          *gg.Context
	Points           float64
	Width            float64
	Value            string
	HorizontalMargin float64
}
