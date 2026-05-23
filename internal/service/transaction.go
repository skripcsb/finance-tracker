package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/nastya/finance-tracker/internal/model"
	"github.com/nastya/finance-tracker/internal/storage"
)

type TransactionService struct {
	txns       *storage.TransactionStorage
	categories *storage.CategoryStorage
}

func NewTransactionService(txns *storage.TransactionStorage, categories *storage.CategoryStorage) *TransactionService {
	return &TransactionService{txns: txns, categories: categories}
}

func (s *TransactionService) Create(ctx context.Context, userID int, req model.CreateTransactionRequest) (model.Transaction, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}

	t := model.Transaction{
		UserID:     userID,
		Type:       req.Type,
		Amount:     req.Amount,
		CategoryID: req.CategoryID,
		Note:       req.Note,
		Date:       date,
	}

	return s.txns.Create(ctx, t)
}

func (s *TransactionService) List(ctx context.Context, userID int, filter model.TransactionFilter) ([]model.Transaction, error) {
	if filter.PerPage <= 0 {
		filter.PerPage = 20
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	return s.txns.List(ctx, userID, filter)
}

func (s *TransactionService) Report(ctx context.Context, userID int, from, to string) ([]model.ReportItem, error) {
	return s.txns.Report(ctx, userID, from, to)
}

func (s *TransactionService) ExportCSV(ctx context.Context, userID int, filter model.TransactionFilter, w io.Writer) error {
	filter.PerPage = 0 // без лимита для экспорта

	txns, err := s.txns.List(ctx, userID, filter)
	if err != nil {
		return err
	}

	csvW := csv.NewWriter(w)
	csvW.Write([]string{"ID", "Type", "Amount", "Category", "Note", "Date"})

	for _, t := range txns {
		csvW.Write([]string{
			fmt.Sprintf("%d", t.ID),
			string(t.Type),
			fmt.Sprintf("%.2f", t.Amount),
			t.Category,
			t.Note,
			t.Date.Format("2006-01-02"),
		})
	}

	csvW.Flush()
	return csvW.Error()
}

func (s *TransactionService) CreateCategory(ctx context.Context, userID int, name string) (model.Category, error) {
	return s.categories.Create(ctx, userID, name)
}

func (s *TransactionService) ListCategories(ctx context.Context, userID int) ([]model.Category, error) {
	return s.categories.ListByUser(ctx, userID)
}
