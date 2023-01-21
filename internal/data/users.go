package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var AnonymousUser = &User{}

type UserModelInterface interface {
	GetByID(id uuid.UUID) (*User, error)
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Activated bool      `json:"activated"`
	Version   int       `json:"version"`
}

type UserModel struct {
	DB *sql.DB
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (m UserModel) GetByID(id uuid.UUID) (*User, error) {
	query := `
        SELECT id, created_at, email, first_name, last_name, activated, version
        FROM users
        WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Email,
		&user.FirstName,
		&user.LastName,
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
