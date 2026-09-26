package ports

import (
	"context"

	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type CategoryBudgetRepository interface {
	Save(ctx context.Context, budget money.CategoryBudget) error
	SetDailyDisplay(ctx context.Context, userID money.UserID, categoryCode string, enabled bool) error
	List(ctx context.Context, userID money.UserID) ([]money.CategoryBudget, error)
	Delete(ctx context.Context, userID money.UserID, categoryCode string) error
}
