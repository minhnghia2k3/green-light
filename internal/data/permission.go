package data

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	"time"
)

type Permissions []string

// Include helper method to check whether Permissions slice contains a specific permission code
func (p Permissions) Include(code string) bool {
	for i := range p {
		if p[i] == code {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *sql.DB
}

// The GetAllForUser method returns all permissions codes for a specific user in a Permission slice.
func (m *PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
			SELECT permissions.code
			FROM permissions
			INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
			INNER JOIN users ON users_permissions.user_id = users.id
			WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan rows into permissions slice
	var permissions Permissions
	for rows.Next() {
		var permission string
		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// The AddForUser method will add permission codes to a user
func (m *PermissionModel) AddForUser(userID int64, code ...string) error {
	// SELECT an temp table with corresponding userID and permission codes
	// then INSERT INTO users_permission table
	query := `
	INSERT INTO users_permissions
	SELECT $1, permissions.id FROM permissions
	WHERE permissions.code = ANY($2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{userID, pq.Array(code)}

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
