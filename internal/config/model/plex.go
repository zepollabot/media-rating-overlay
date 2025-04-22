package config

import "fmt"

type Plex struct {
	Url       string    `yaml:"url"`
	Token     string    `yaml:"token"`
	Enabled   bool      `yaml:"enabled"`
	Libraries []Library `yaml:"libraries"`
}

func DefaultPlex() *Plex {
	return &Plex{
		Enabled:   false,
		Libraries: []Library{},
	}
}

// Validate validates the Plex configuration
func (c *Plex) Validate() error {
	if c.Enabled {
		if c.Url == "" {
			return fmt.Errorf("plex.url is required when plex is enabled")
		}
		if c.Token == "" {
			return fmt.Errorf("plex.token is required when plex is enabled")
		}
	}
	return nil
}
