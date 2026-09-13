package queries

import (
	"context"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type MonthlySummaryQuery struct {
	UserID string
	Now    time.Time
}

type MonthlySummaryResult struct {
	From         time.Time
	To           time.Time
	IncomeCents  int64
	ExpenseCents int64
	BalanceCents int64
}

type MonthlySummaryHandler struct {
	reader ports.TransactionReader
}

func NewMonthlySummaryHandler(reader ports.TransactionReader) MonthlySummaryHandler {
	return MonthlySummaryHandler{reader: reader}
}

func (h MonthlySummaryHandler) Handle(ctx context.Context, query MonthlySummaryQuery) (MonthlySummaryResult, error) {
	now := query.Now
	if now.IsZero() {
		now = time.Now()
	}
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, 0)

	userID, err := money.NewUserID(query.UserID)
	if err != nil {
		return MonthlySummaryResult{}, err
	}
	income, expense, err := h.reader.SumByType(ctx, userID, from, to)
	if err != nil {
		return MonthlySummaryResult{}, err
	}
	return MonthlySummaryResult{
		From:         from,
		To:           to,
		IncomeCents:  income,
		ExpenseCents: expense,
		BalanceCents: income - expense,
	}, nil
}
