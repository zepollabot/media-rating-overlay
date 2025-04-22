package config

type Overlay struct {
	Type         string  `yaml:"type"`
	Height       float64 `yaml:"height"`
	Transparency float64 `yaml:"transparency"`
}
