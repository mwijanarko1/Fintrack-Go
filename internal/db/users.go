package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"fintrack-go/internal/models"
)

func (db *DB) CreateUser(ctx context.Context, email string) (*models.User, error) {
	query := `INSERT INTO users (email) VALUES ($1) RETURNING id, email, created_at`
	
	var user models.User
	err := db.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return nil, ErrDuplicateEmail
			}
		}
		return nil, err
	}
	
	return &user, nil
}

func (db *DB) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, email, created_at FROM users WHERE id = $1`
	
	var user models.User
	err := db.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, created_at FROM users WHERE email = $1`
	
	var user models.User
	err := db.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}
