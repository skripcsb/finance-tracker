package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vadim/finance-tracker/internal/model"
)

var ErrCategoryExists = errors.New("category already exists")

type CategoryStorage struct {
	pool *pgxpool.Pool
}

func NewCategoryStorage(pool *pgxpool.Pool) *CategoryStorage {
	return &CategoryStorage{pool: pool}
}

func (s *CategoryStorage) Create(ctx context.Context, userID int, name string) (model.Category, error) {
	var cat model.Category
	err := s.pool.QueryRow(ctx,
		`INSERT INTO categories (user_id, name) VALUES ($1, $2) RETURNING id, user_id, name`,
		userID, name,
	).Scan(&cat.ID, &cat.UserID, &cat.Name)

	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return model.Category{}, ErrCategoryExists
	}
	return cat, err
}

func (s *CategoryStorage) ListByUser(ctx context.Context, userID int) ([]model.Category, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name FROM categories WHERE user_id = $1 ORDER BY name`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (s *CategoryStorage) CreateDefaults(ctx context.Context, userID int) error {
	defaults := []string{"Еда", "Транспорт", "Жильё", "Развлечения", "Зарплата", "Другое"}
	for _, name := range defaults {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO categories (user_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			userID, name,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
