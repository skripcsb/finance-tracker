package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vadim/finance-tracker/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")

type UserStorage struct {
	db *pgxpool.Pool
}

func NewUserStorage(db *pgxpool.Pool) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) Create(ctx context.Context, username, passHash string) (model.User, error) {
	row := s.db.QueryRow(ctx,
		`INSERT INTO users (username, password) VALUES ($1, $2)
		 RETURNING id, username, created_at`,
		username, passHash,
	)

	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.CreatedAt)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return model.User{}, ErrUserExists
	}
	return u, err
}

func (s *UserStorage) GetByUsername(ctx context.Context, username string) (model.User, error) {
	var u model.User
	err := s.db.QueryRow(ctx,
		`SELECT id, username, password, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	return u, err
}
