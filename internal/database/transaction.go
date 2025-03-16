package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"accountant/internal/models"
)

// Initialize transaction table
func (s *service) InitTransactionTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			description TEXT NOT NULL,
			amount REAL NOT NULL,
			category TEXT NOT NULL,
			notes TEXT,
			tags TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`

	_, err := s.db.Exec(query)
	return err
}


func (s *service) CreateTransaction(ctx context.Context, input models.TransactionInput) (*models.Transaction, error) {
	if !models.ValidCategories[input.Category] {
		return nil, errors.New("invalid category")
	}

	now := time.Now()
	tagJSON, err := json.Marshal(input.Tags) // Mappping input.Tags to JSON
	if err != nil {
		return nil, fmt.Errorf("error marshaling tags: %w", err)
	}

	query := `
		INSERT INTO transactions (date, description, amount, category, notes, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, date, description, amount, category, notes, tags, created_at, updated_at
	`

	var transaction models.Transaction
	var tagsStr string

	err = s.db.QueryRowContext(
		ctx,
		query,
		input.Date.Format(time.RFC3339),
		input.Description,
		input.Amount,
		input.Category,
		input.Notes,
		string(tagJSON),
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	).Scan(
		&transaction.ID,
		&transaction.Date,
		&transaction.Description,
		&transaction.Amount,
		&transaction.Category,
		&transaction.Notes,
		&tagsStr,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}

	// Parse tags
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &transaction.Tags); err != nil {
			transaction.Tags = []string{}
		}
	}

	return &transaction, nil
}


func (s *service) GetTransaction(ctx context.Context, id int64) (*models.Transaction, error) {
	query := `
		SELECT id, date, description, amount, category, notes, tags, created_at, updated_at
		FROM transactions
		WHERE id = ?
	`

	var transaction models.Transaction
	var tagsStr string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.Date,
		&transaction.Description,
		&transaction.Amount,
		&transaction.Category,
		&transaction.Notes,
		&tagsStr,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaction not found: %d", id)
		}
		return nil, fmt.Errorf("error retrieving transaction: %w", err)
	}

	// Parse tags
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &transaction.Tags); err != nil {
			transaction.Tags = []string{}
		}
	}

	return &transaction, nil
}

func (s *service) UpdateTransaction(ctx context.Context, id int64, input models.TransactionInput) (*models.Transaction, error) {
	if !models.ValidCategories[input.Category] {
		return nil, errors.New("invalid category")
	}

	now := time.Now()
	tagJSON, err := json.Marshal(input.Tags)
	if err != nil {
		return nil, fmt.Errorf("error marshaling tags: %w", err)
	}

	query := `
		UPDATE transactions
		SET date = ?, description = ?, amount = ?, category = ?, notes = ?, tags = ?, updated_at = ?
		WHERE id = ?
		RETURNING id, date, description, amount, category, notes, tags, created_at, updated_at
	`

	var transaction models.Transaction
	var tagsStr string

	err = s.db.QueryRowContext(
		ctx,
		query,
		input.Date.Format(time.RFC3339),
		input.Description,
		input.Amount,
		input.Category,
		input.Notes,
		string(tagJSON),
		now.Format(time.RFC3339),
		id,
	).Scan(
		&transaction.ID,
		&transaction.Date,
		&transaction.Description,
		&transaction.Amount,
		&transaction.Category,
		&transaction.Notes,
		&tagsStr,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaction not found: %d", id)
		}
		return nil, fmt.Errorf("error updating transaction: %w", err)
	}

	// Parse tags
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &transaction.Tags); err != nil {
			transaction.Tags = []string{}
		}
	}

	return &transaction, nil
}

// DeleteTransaction removes a transaction by ID
func (s *service) DeleteTransaction(ctx context.Context, id int64) error {
	query := `DELETE FROM transactions WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found: %d", id)
	}

	return nil
}

// ListTransactions retrieves transactions with optional filters
func (s *service) ListTransactions(ctx context.Context, filters map[string]string) ([]*models.Transaction, error) {
	baseQuery := `
		SELECT id, date, description, amount, category, notes, tags, created_at, updated_at
		FROM transactions
	`

	// Build where clause and args
	var whereConditions []string
	var args []interface{}

	if category, ok := filters["category"]; ok && category != "" {
		whereConditions = append(whereConditions, "category = ?")
		args = append(args, category)
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		whereConditions = append(whereConditions, "date >= ?")
		args = append(args, startDate)
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		whereConditions = append(whereConditions, "date <= ?")
		args = append(args, endDate)
	}

	// Construct the full query
	query := baseQuery
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	query += " ORDER BY date DESC"

	// Add optional limit
	if limit, ok := filters["limit"]; ok && limit != "" {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying transactions: %w", err)
	}
	defer rows.Close()

	// Process results
	var transactions []*models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		var tagsStr string

		err := rows.Scan(
			&transaction.ID,
			&transaction.Date,
			&transaction.Description,
			&transaction.Amount,
			&transaction.Category,
			&transaction.Notes,
			&tagsStr,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning transaction: %w", err)
		}

		// Parse tags
		if tagsStr != "" {
			if err := json.Unmarshal([]byte(tagsStr), &transaction.Tags); err != nil {
				transaction.Tags = []string{}
			}
		}

		transactions = append(transactions, &transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}