package db

import (
	"context"
	"strconv"
	"time"

	"fintrack-go/internal/models"
)

func (db *DB) GetSummary(ctx context.Context, userID string, from, to *time.Time) (*models.Summary, error) {
	query := `
		SELECT 
			COALESCE(c.id, NULL) as category_id,
			COALESCE(c.name, 'Uncategorized') as category_name,
			COALESCE(SUM(t.amount), 0) as total
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
	
	query += ` GROUP BY c.id, c.name ORDER BY category_name`
	
	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var categories []models.CategorySummary
	for rows.Next() {
		var summary models.CategorySummary
		if err := rows.Scan(&summary.CategoryID, &summary.CategoryName, &summary.Total); err != nil {
			return nil, err
		}
		categories = append(categories, summary)
	}
	
	now := time.Now()
	defaultFrom := now.AddDate(0, 0, -30)
	defaultTo := now
	
	fromTime := defaultFrom
	toTime := defaultTo
	
	if from != nil {
		fromTime = *from
	}
	if to != nil {
		toTime = *to
	}
	
	summary := &models.Summary{
		UserID:    userID,
		From:      fromTime,
		To:        toTime,
		Categories: categories,
	}
	
	return summary, nil
}
