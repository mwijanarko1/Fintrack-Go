package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"fintrack-go/internal/models"
)

func (db *DB) CreateCategory(ctx context.Context, userID, name string) (*models.Category, error) {
	query := `INSERT INTO categories (user_id, name) VALUES ($1, $2) RETURNING id, user_id, name, created_at`
	
	var category models.Category
	err := db.pool.QueryRow(ctx, query, userID, name).Scan(&category.ID, &category.UserID, &category.Name, &category.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "categories_user_id_name_key" {
					return nil, ErrDuplicateCategory
				}
			}
			if pgErr.Code == "23503" {
				return nil, ErrUserNotFound
			}
		}
		return nil, err
	}
	
	return &category, nil
}

func (db *DB) ListCategories(ctx context.Context, userID string) ([]models.Category, error) {
	query := `SELECT id, user_id, name, created_at FROM categories WHERE user_id = $1 ORDER BY created_at DESC`
	
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.UserID, &category.Name, &category.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	
	return categories, nil
}

func (db *DB) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	query := `SELECT id, user_id, name, created_at FROM categories WHERE id = $1`
	
	var category models.Category
	err := db.pool.QueryRow(ctx, query, id).Scan(&category.ID, &category.UserID, &category.Name, &category.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return &category, nil
}
