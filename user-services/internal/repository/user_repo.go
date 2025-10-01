package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        int64
	Email     string
	Name      string
	Password  string // bcrypt hash
	CreatedAt time.Time
}


// CreateUser inserts a new user with unique email.
func CreateUser(ctx context.Context, db *pgxpool.Pool, email, name, passwordHash string) (User, error) {
	var u User
	row := db.QueryRow(ctx, `
		INSERT INTO users (email, name, password)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, password, created_at;
	`, email, name, passwordHash)
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt); err != nil {
		return u, err
	}
	return u, nil
}
