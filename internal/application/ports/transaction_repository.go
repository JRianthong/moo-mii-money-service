package ports

import (
	"context"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type TransactionWriter interface {
	Save(ctx context.Context, tx money.Transaction) (SaveTransactionResult, error)
}

type SaveTransactionResult struct {
	Transaction     money.Transaction
	AlreadyRecorded bool
}

type TransactionReader interface {
	SumByType(ctx context.Context, userID money.UserID, from, to time.Time) (incomeCents int64, expenseCents int64, err error)
	ListRecent(ctx context.Context, userID money.UserID, limit int) ([]money.Transaction, error)
}
