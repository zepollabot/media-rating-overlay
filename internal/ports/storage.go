package ports

type PosterStorage interface {
	CheckIfPosterExists(path string) (bool, error)
	SavePoster(path string, data []byte) error
}
