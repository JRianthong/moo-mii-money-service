package queries

import (
	"context"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type ListRecentTransactionsQuery struct {
	UserID string
	Limit  int
}

type ListRecentTransactionsHandler struct {
	reader ports.TransactionReader
}

func NewListRecentTransactionsHandler(reader ports.TransactionReader) ListRecentTransactionsHandler {
	return ListRecentTransactionsHandler{reader: reader}
}

func (h ListRecentTransactionsHandler) Handle(ctx context.Context, query ListRecentTransactionsQuery) ([]money.Transaction, error) {
	limit := query.Limit
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	userID, err := money.NewUserID(query.UserID)
	if err != nil {
		return nil, err
	}
	return h.reader.ListRecent(ctx, userID, limit)
}
