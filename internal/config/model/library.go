package config

type Library struct {
	Name    string  `yaml:"name"`
	Enabled bool    `yaml:"enabled"`
	Refresh bool    `yaml:"refresh"`
	Path    string  `yaml:"path"`
	Filters Filter  `yaml:"filters"`
	Overlay Overlay `yaml:"overlay"`
}
