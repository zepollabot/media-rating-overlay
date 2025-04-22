package model

// PosterError represents an error in poster processing
type PosterError struct {
	Stage string
	Err   error
}

func (e *PosterError) Error() string {
	if e.Stage == "" {
		return "Invalid stage in PosterError"
	}
	if e.Err == nil {
		return e.Stage + ": Invalid error in PosterError"
	}
	return e.Stage + ": " + e.Err.Error()
}
