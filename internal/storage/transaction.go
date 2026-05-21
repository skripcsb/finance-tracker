package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vadim/finance-tracker/internal/model"
)

type TransactionStorage struct {
	pool *pgxpool.Pool
}

func NewTransactionStorage(pool *pgxpool.Pool) *TransactionStorage {
	return &TransactionStorage{pool: pool}
}

func (s *TransactionStorage) Create(ctx context.Context, t model.Transaction) (model.Transaction, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO transactions (user_id, type, amount, category_id, note, date)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`,
		t.UserID, t.Type, t.Amount, t.CategoryID, t.Note, t.Date,
	).Scan(&t.ID, &t.CreatedAt)
	return t, err
}

func (s *TransactionStorage) List(ctx context.Context, userID int, f model.TransactionFilter) ([]model.Transaction, error) {
	query := `SELECT t.id, t.user_id, t.type, t.amount, t.category_id, c.name, t.note, t.date, t.created_at
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1`
	args := []any{userID}
	argIdx := 2

	if f.Type != "" {
		query += fmt.Sprintf(` AND t.type = $%d`, argIdx)
		args = append(args, string(f.Type))
		argIdx++
	}
	if f.CategoryID > 0 {
		query += fmt.Sprintf(` AND t.category_id = $%d`, argIdx)
		args = append(args, f.CategoryID)
		argIdx++
	}
	if f.From != "" {
		query += fmt.Sprintf(` AND t.date >= $%d`, argIdx)
		args = append(args, f.From)
		argIdx++
	}
	if f.To != "" {
		query += fmt.Sprintf(` AND t.date <= $%d`, argIdx)
		args = append(args, f.To)
		argIdx++
	}

	query += ` ORDER BY t.date DESC`

	if f.PerPage > 0 {
		offset := (f.Page - 1) * f.PerPage
		if offset < 0 {
			offset = 0
		}
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, f.PerPage, offset)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]model.Transaction, 0)
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.CategoryID, &t.Category, &t.Note, &t.Date, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

// TODO: добавить группировку по type (отдельно доходы/расходы)
func (s *TransactionStorage) Report(ctx context.Context, userID int, from, to string) ([]model.ReportItem, error) {
	query := `SELECT c.name, SUM(t.amount) as total
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1`
	args := []any{userID}
	argIdx := 2

	var conditions []string
	if from != "" {
		conditions = append(conditions, fmt.Sprintf("t.date >= $%d", argIdx))
		args = append(args, from)
		argIdx++
	}
	if to != "" {
		conditions = append(conditions, fmt.Sprintf("t.date <= $%d", argIdx))
		args = append(args, to)
		argIdx++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += ` GROUP BY c.name ORDER BY total DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ReportItem
	for rows.Next() {
		var item model.ReportItem
		if err := rows.Scan(&item.Category, &item.Total); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
