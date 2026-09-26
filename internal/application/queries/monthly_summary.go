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
	reader   ports.TransactionReader
	settings ports.UserSettingsRepository
}

func NewMonthlySummaryHandler(reader ports.TransactionReader, settings ports.UserSettingsRepository) MonthlySummaryHandler {
	return MonthlySummaryHandler{reader: reader, settings: settings}
}

func (h MonthlySummaryHandler) Handle(ctx context.Context, query MonthlySummaryQuery) (MonthlySummaryResult, error) {
	now := query.Now
	if now.IsZero() {
		now = time.Now()
	}
	userID, err := money.NewUserID(query.UserID)
	if err != nil {
		return MonthlySummaryResult{}, err
	}
	cycle, err := h.settings.BillingCycle(ctx, userID)
	if err != nil {
		return MonthlySummaryResult{}, err
	}
	from, to := cycle.Period(now)
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
