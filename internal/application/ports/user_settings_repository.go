package ports

import (
	"context"

	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type UserSettingsRepository interface {
	BillingCycle(ctx context.Context, userID money.UserID) (money.BillingCycle, error)
	SetBillingCycle(ctx context.Context, userID money.UserID, cycle money.BillingCycle) error
}
