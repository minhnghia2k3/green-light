package data

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
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

// MovieModel struct which wrap a sql.DB connection pool
type MovieModel struct {
	DB *sql.DB
}

func ValidateMovie(v *validation.Validator, input *Movie) {
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

func (m *MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`

	// Wrap input into []args
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Query a row then scan value into destination
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)

}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, runtime, genres, version FROM movies
		WHERE id = $1
	`

	var movie Movie

	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m *MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1,year = $2,runtime= $3,genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version
	`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID}

	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

func (m *MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM movies
		WHERE id = $1
	`

	results, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
