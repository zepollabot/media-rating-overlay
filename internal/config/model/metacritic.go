package config

import "fmt"

// Metacritic configuration
type Metacritic struct {
	Enabled bool   `yaml:"enabled"`
	APIKey  string `yaml:"api_key"`
}

func DefaultMetacritic() *Metacritic {
	return &Metacritic{
		Enabled: false,
	}
}

func (c *Metacritic) Validate() error {
	if c.Enabled {
		if c.APIKey == "" {
			return fmt.Errorf("metacritic.api_key is required when metacritic is enabled")
		}
	}
	return nil
}
