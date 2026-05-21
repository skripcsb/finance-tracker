package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vadim/finance-tracker/internal/auth"
)

func SetupRoutes(authH *AuthHandler, txnH *TransactionHandler, tm *auth.TokenManager, log *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(logRequests(log))

	// публичные
	r.Post("/api/register", authH.Register)
	r.Post("/api/login", authH.Login)

	// защищённые
	r.Group(func(r chi.Router) {
		r.Use(tm.RequireAuth)

		r.Post("/api/transactions", txnH.Create)
		r.Get("/api/transactions", txnH.List)
		r.Get("/api/reports", txnH.Report)
		r.Get("/api/export/csv", txnH.ExportCSV)
		r.Post("/api/categories", txnH.CreateCategory)
		r.Get("/api/categories", txnH.ListCategories)
	})

	return r
}

func logRequests(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t := time.Now()
			next.ServeHTTP(w, r)
			log.Info("http", "method", r.Method, "path", r.URL.Path, "took", time.Since(t))
		})
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
