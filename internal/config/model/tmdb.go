package config

import "fmt"

type TMDB struct {
	Enabled  bool   `yaml:"enabled"`
	ApiKey   string `yaml:"api_key"`
	Language string `yaml:"language"`
	Region   string `yaml:"region"`
}

func DefaultTMDB() *TMDB {
	return &TMDB{
		Enabled:  false,
		Language: "en-US",
		Region:   "US",
	}
}

// Validate validates the TMDB configuration
func (c *TMDB) Validate() error {
	if c.Enabled {
		if c.ApiKey == "" {
			return fmt.Errorf("tmdb.api_key is required when tmdb is enabled")
		}
	}
	return nil
}
