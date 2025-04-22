package model

import "path/filepath"

// PosterConfig contains configuration for poster generation
type PosterConfig struct {
	Dimensions struct {
		Width  int
		Height int
	}
	Margins struct {
		Left  int
		Right int
	}
	ImagePaths struct {
		RottenTomatoes struct {
			Critic struct {
				Normal string
				Low    string
			}
			Audience struct {
				Normal string
				Low    string
			}
		}
		IMDB struct {
			Audience struct {
				Normal string
			}
		}
		TMDB struct {
			Audience struct {
				Normal string
			}
		}
	}
	VisualDebug bool
}

func PosterConfigWithDefaultValues() *PosterConfig {
	config := &PosterConfig{}

	config.Dimensions.Width = 1200
	config.Dimensions.Height = 1800

	config.Margins.Left = 20
	config.Margins.Right = 20

	config.ImagePaths.RottenTomatoes.Critic.Normal = filepath.Join("internal", "processor", "image", "data", "RT_critic.png")
	config.ImagePaths.RottenTomatoes.Critic.Low = filepath.Join("internal", "processor", "image", "data", "RT_critic_low.png")
	config.ImagePaths.RottenTomatoes.Audience.Normal = filepath.Join("internal", "processor", "image", "data", "RT_audience.png")
	config.ImagePaths.RottenTomatoes.Audience.Low = filepath.Join("internal", "processor", "image", "data", "RT_audience_low.png")
	config.ImagePaths.IMDB.Audience.Normal = filepath.Join("internal", "processor", "image", "data", "IMDb.png")
	config.ImagePaths.TMDB.Audience.Normal = filepath.Join("internal", "processor", "image", "data", "TMDB.png")

	return config
}
