package commands

import (
	"context"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type SetBillingCycleCommand struct {
	UserID   string
	StartDay int
}

type SetBillingCycleHandler struct {
	settings ports.UserSettingsRepository
}

func NewSetBillingCycleHandler(settings ports.UserSettingsRepository) SetBillingCycleHandler {
	return SetBillingCycleHandler{settings: settings}
}

func (h SetBillingCycleHandler) Handle(ctx context.Context, command SetBillingCycleCommand) (money.BillingCycle, error) {
	userID, err := money.NewUserID(command.UserID)
	if err != nil {
		return money.BillingCycle{}, err
	}
	cycle, err := money.NewBillingCycle(command.StartDay)
	if err != nil {
		return money.BillingCycle{}, err
	}
	if err := h.settings.SetBillingCycle(ctx, userID, cycle); err != nil {
		return money.BillingCycle{}, err
	}
	return cycle, nil
}
