package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"version"`
}

type password struct {
	plaintext *string
	hash      []byte
}

type UserModel struct {
	DB *sql.DB
}

// The Set method calculates the bcrypt hash of a plaintext password,
// and stores both the hash and the plaintext versions in the struct.
func (p *password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plainTextPassword
	p.hash = hash
	return nil
}

// Matches method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *password) Matches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validation.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validation.Matches(email, validation.EmailRx), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validation.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validation.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// Insert a new record in the database for the user.
func (m UserModel) Insert(user *User) error {

	query := `INSERT INTO users (name, email, hashed_password, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil

}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `SELECT id, created_at, name, email, hashed_password,activated, version
FROM users
WHERE email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update the details for a specific user. Notice that we check against
// the version field to help prevent `race conditions`
func (m UserModel) Update(user *User) error {
	query := `UPDATE users
	SET name = $1, email = $2, hashed_password = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version
	`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrConflictEdit
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetForToken(scope string, tokenPlainText string) (*User, error) {
	var user User
	// Calculate SHA-256 hash from the plaintext token
	tokenHash := sha256.Sum256([]byte(tokenPlainText)) // return an array with length 32

	query := `
	SELECT users.id, users.created_at, users.name, users.email, users.hashed_password, users.activated, users.version
	FROM users
	INNER JOIN tokens
	ON users.id = tokens.user_id
	WHERE tokens.hash = $1 AND tokens.scope = $2 AND tokens.expiry > $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{tokenHash[:], scope, time.Now()}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
