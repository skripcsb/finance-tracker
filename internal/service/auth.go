package service

import (
	"context"
	"errors"

	"github.com/vadim/finance-tracker/internal/auth"
	"github.com/vadim/finance-tracker/internal/model"
	"github.com/vadim/finance-tracker/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var ErrWrongCredentials = errors.New("wrong username or password")

type AuthService struct {
	users      *storage.UserStorage
	categories *storage.CategoryStorage
	tokens     *auth.TokenManager
}

func NewAuthService(users *storage.UserStorage, categories *storage.CategoryStorage, tokens *auth.TokenManager) *AuthService {
	return &AuthService{users: users, categories: categories, tokens: tokens}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	user, err := s.users.Create(ctx, req.Username, string(hashed))
	if err != nil {
		return model.User{}, err
	}

	s.categories.CreateDefaults(ctx, user.ID)

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (string, error) {
	user, err := s.users.GetByUsername(ctx, req.Username)
	if err != nil {
		if err == storage.ErrUserNotFound {
			return "", ErrWrongCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", ErrWrongCredentials
	}

	return s.tokens.Issue(user.ID)
}
