package data

import (
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"time"
)

/* Fields are capital, which is necessary for encoding/json package*/

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime"` // <- Custom Runtime `type`
	Genres    []string  `json:"genres"`
	Version   int32     `json:"version"`
}

func ValidateMovies(v *validation.Validator, input *Movie) {
	v.Check(input.Title != "", "input", "must be provided")
	v.Check(len(input.Title) <= 500, "input", "must not be more than 500 bytes long")

	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime >= 0, "runtime", "must not be negative")

	v.Check(input.Genres != nil, "genres", "must be provided")
	v.Check(len(input.Genres) >= 1, "genres", "must contains at least 1 genre")
	v.Check(len(input.Genres) <= 5, "genres", "must not contain more than 5 genre")
	v.Check(validation.Unique(input.Genres), "genres", "must not contain duplicate values")
}
