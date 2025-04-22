package model

type Logo struct {
	Image    Image
	Text     Text
	SumWidth int
}

// LogoDimensions represents the dimensions for a logo
type LogoDimensions struct {
	AreaWidth  float64
	AreaHeight float64
	FontSize   float64
}
