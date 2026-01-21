package db

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"fintrack-go/internal/models"
)

func (db *DB) CreateTransaction(ctx context.Context, userID string, categoryID *string, amount float64, description *string, occurredAt time.Time) (*models.Transaction, error) {
	if categoryID != nil {
		if err := db.ValidateCategoryOwnership(ctx, *categoryID, userID); err != nil {
			return nil, errors.New("category does not belong to user")
		}
	}
	
	query := `INSERT INTO transactions (user_id, category_id, amount, description, occurred_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, category_id, amount, description, occurred_at, created_at`
	
	var transaction models.Transaction
	err := db.pool.QueryRow(ctx, query, userID, categoryID, amount, description, occurredAt).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.CategoryID,
		&transaction.Amount,
		&transaction.Description,
		&transaction.OccurredAt,
		&transaction.CreatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23503" {
				return nil, errors.New("user not found")
			}
		}
		return nil, err
	}
	
	return &transaction, nil
}

func (db *DB) ListTransactions(ctx context.Context, userID string, from, to *time.Time) ([]models.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.category_id, t.amount, t.description, t.occurred_at, t.created_at,
			c.name as category_name
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1
	`
	args := []interface{}{userID}
	argCount := 1
	
	if from != nil {
		argCount++
		query += ` AND t.occurred_at >= $` + strconv.Itoa(argCount)
		args = append(args, *from)
	}
	
	if to != nil {
		argCount++
		query += ` AND t.occurred_at <= $` + strconv.Itoa(argCount)
		args = append(args, *to)
	}
	
	query += ` ORDER BY t.occurred_at DESC`
	
	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		var categoryName *string
		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.CategoryID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.OccurredAt,
			&transaction.CreatedAt,
			&categoryName,
		); err != nil {
			return nil, err
		}
		transaction.CategoryName = categoryName
		transactions = append(transactions, transaction)
	}
	
	return transactions, nil
}

func (db *DB) GetTransactionByID(ctx context.Context, id string) (*models.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.category_id, t.amount, t.description, t.occurred_at, t.created_at,
			c.name as category_name
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.id = $1
	`
	
	var transaction models.Transaction
	var categoryName *string
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.CategoryID,
		&transaction.Amount,
		&transaction.Description,
		&transaction.OccurredAt,
		&transaction.CreatedAt,
		&categoryName,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrTransactionNotFound
	}
	if err != nil {
		return nil, err
	}
	transaction.CategoryName = categoryName
	
	return &transaction, nil
}

func (db *DB) ValidateCategoryOwnership(ctx context.Context, categoryID, userID string) error {
	query := `SELECT 1 FROM categories WHERE id = $1 AND user_id = $2`
	
	var exists bool
	err := db.pool.QueryRow(ctx, query, categoryID, userID).Scan(&exists)
	if err == pgx.ErrNoRows {
		return errors.New("category does not belong to user")
	}
	if err != nil {
		return err
	}
	
	return nil
}
