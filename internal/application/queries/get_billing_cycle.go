package queries

import (
	"context"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type GetBillingCycleHandler struct{ settings ports.UserSettingsRepository }

func NewGetBillingCycleHandler(settings ports.UserSettingsRepository) GetBillingCycleHandler {
	return GetBillingCycleHandler{settings: settings}
}

func (h GetBillingCycleHandler) Handle(ctx context.Context, userID string) (money.BillingCycle, error) {
	user, err := money.NewUserID(userID)
	if err != nil {
		return money.BillingCycle{}, err
	}
	return h.settings.BillingCycle(ctx, user)
}
