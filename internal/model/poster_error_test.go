package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PosterErrorTestSuite struct {
	suite.Suite
}

func TestPosterErrorTestSuite(t *testing.T) {
	suite.Run(t, new(PosterErrorTestSuite))
}

func (s *PosterErrorTestSuite) TestPosterError_Error() {
	tests := []struct {
		name        string
		stage       string
		originalErr error
		expectedMsg string
	}{
		{
			name:        "Error with specific stage and non-nil error",
			stage:       "ImageLoading",
			originalErr: errors.New("file not found"),
			expectedMsg: "ImageLoading: file not found",
		},
		{
			name:        "Error with empty stage and non-nil error",
			stage:       "",
			originalErr: errors.New("some generic error"),
			expectedMsg: "Invalid stage in PosterError",
		},
		{
			name:        "Error with nil original error",
			stage:       "Validation",
			originalErr: nil,
			expectedMsg: "Validation: Invalid error in PosterError",
		},
		{
			name:        "Error with nil original error and empty stage",
			stage:       "",
			originalErr: nil,
			expectedMsg: "Invalid stage in PosterError",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			posterErr := &PosterError{Stage: tt.stage, Err: tt.originalErr}
			assert.Equal(t, tt.expectedMsg, posterErr.Error())
		})
	}
}
