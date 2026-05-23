package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/nastya/finance-tracker/internal/auth"
	"github.com/nastya/finance-tracker/internal/model"
	"github.com/nastya/finance-tracker/internal/service"
	"github.com/nastya/finance-tracker/internal/storage"
)

type TransactionHandler struct {
	svc *service.TransactionService
	log *slog.Logger
}

func NewTransactionHandler(svc *service.TransactionService, log *slog.Logger) *TransactionHandler {
	return &TransactionHandler{svc: svc, log: log}
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r.Context())

	var req model.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "amount must be positive"})
		return
	}
	if req.Type != model.Income && req.Type != model.Expense {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type must be income or expense"})
		return
	}

	txn, err := h.svc.Create(r.Context(), uid, req)
	if err != nil {
		h.log.Error("create txn", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		return
	}

	writeJSON(w, http.StatusCreated, txn)
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r.Context())
	q := r.URL.Query()

	filter := model.TransactionFilter{
		Type: model.TransactionType(q.Get("type")),
		From: q.Get("from"),
		To:   q.Get("to"),
	}
	if v := q.Get("category_id"); v != "" {
		n, _ := strconv.Atoi(v)
		filter.CategoryID = n
	}
	if v := q.Get("page"); v != "" {
		n, _ := strconv.Atoi(v)
		filter.Page = n
	}
	if v := q.Get("per_page"); v != "" {
		n, _ := strconv.Atoi(v)
		filter.PerPage = n
	}

	txns, err := h.svc.List(r.Context(), uid, filter)
	if err != nil {
		h.log.Error("list txns", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		return
	}
	writeJSON(w, http.StatusOK, txns)
}

func (h *TransactionHandler) Report(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r.Context())

	items, err := h.svc.Report(r.Context(), uid, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		h.log.Error("report", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *TransactionHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r.Context())

	filter := model.TransactionFilter{
		From: r.URL.Query().Get("from"),
		To:   r.URL.Query().Get("to"),
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="transactions.csv"`)

	if err := h.svc.ExportCSV(r.Context(), uid, filter, w); err != nil {
		// уже начали писать body, не можем вернуть ошибку клиенту
		h.log.Error("csv export", "err", err)
	}
}

func (h *TransactionHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r.Context())

	var req model.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}

	cat, err := h.svc.CreateCategory(r.Context(), uid, req.Name)
	if err == storage.ErrCategoryExists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already exists"})
		return
	}
	if err != nil {
		h.log.Error("create category", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		return
	}

	writeJSON(w, http.StatusCreated, cat)
}

func (h *TransactionHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r.Context())

	cats, err := h.svc.ListCategories(r.Context(), uid)
	if err != nil {
		h.log.Error("list categories", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		return
	}
	writeJSON(w, http.StatusOK, cats)
}
