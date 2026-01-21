package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"fintrack-go/internal/db"
)

const maxRequestBodySize = 1 << 20

func SetupRoutes(logger zerolog.Logger, database db.Database) chi.Router {
	r := chi.NewRouter()

	healthHandler := NewHealthHandler(logger, database)
	userHandler := NewUserHandler(logger, database)
	categoryHandler := NewCategoryHandler(logger, database)
	transactionHandler := NewTransactionHandler(logger, database)
	summaryHandler := NewSummaryHandler(logger, database)

	r.Get("/health", healthHandler.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/users", userHandler.CreateUser)

		r.Route("/categories", func(r chi.Router) {
			r.Post("/", categoryHandler.CreateCategory)
			r.Get("/", categoryHandler.ListCategories)
		})

		r.Route("/transactions", func(r chi.Router) {
			r.Post("/", transactionHandler.CreateTransaction)
			r.Get("/", transactionHandler.ListTransactions)
		})

		r.Get("/summary", summaryHandler.GetSummary)
	})

	return r
}
