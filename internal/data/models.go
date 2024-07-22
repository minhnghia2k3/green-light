package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrConflictEdit   = errors.New("edit conflict")
)

type IModel interface {
	Insert(*Movie) error
	Get(int64) (*Movie, error)
	Update(*Movie) error
	Delete(int64) error
}

// Models struct which is base model
type Models struct {
	Movies MovieModel
}

// NewModels is a constructor
func NewModels(db *sql.DB) *Models {
	return &Models{
		Movies: MovieModel{DB: db},
	}
}
