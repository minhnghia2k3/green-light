package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"time"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// generateToken will create an randomBytes then encode  the byte slice to
// a base-32-encoded string for Plaintext field, generate SHA-256 hash of the Plaintext
// and store into `tokens` table.
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {

	// Create a Token instance
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte-slice with length of 16 bytes
	randomBytes := make([]byte, 16)

	// Fill the byte-slice with randomBytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string and assign to
	// Plaintext field
	// Plaintext: Y3QMGX3PJ3WLRL2YRTQGQ6KRHU
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate an SHA-256 hash of the plaintext token string.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

// ValidateTokenPlainText Check that the plaintext token has been provided and is exactly 26 bytes long?
func ValidateTokenPlainText(v *validation.Validator, tokenPlainText string) {
	v.Check(tokenPlainText != "", "token", "must be provided")
	v.Check(len(tokenPlainText) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

// The New method is a shortcut which creates a new Token struct and then inserts the
// data into tokens table
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// Insert adds the data for a specific token to the tokens table.
func (m TokenModel) Insert(token *Token) error {
	query := `INSERT INTO tokens(hash, user_id,expiry,scope)
	VALUES($1, $2, $3, $4)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAllForUser deletes all tokens for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `DELETE FROM tokens
	WHERE user_id = $1 AND scope = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, scope)
	if err != nil {
		return err
	}
	return nil
}
