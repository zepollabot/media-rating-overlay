package config

type Filter struct {
	Year    []string `yaml:"years"`
	Title   []string `yaml:"titles"`
	AddedAt string   `yaml:"added_at"`
	Genre   []string `yaml:"genres"`
}
