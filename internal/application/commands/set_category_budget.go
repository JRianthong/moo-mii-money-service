package commands

import (
	"context"
	"errors"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type SetCategoryBudgetCommand struct {
	UserID             string
	Category           string
	MonthlyAmountCents int64
}

type SetCategoryBudgetHandler struct {
	budgets ports.CategoryBudgetRepository
}

func NewSetCategoryBudgetHandler(budgets ports.CategoryBudgetRepository) SetCategoryBudgetHandler {
	return SetCategoryBudgetHandler{budgets: budgets}
}

func (h SetCategoryBudgetHandler) Handle(ctx context.Context, command SetCategoryBudgetCommand) (money.CategoryBudget, error) {
	userID, err := money.NewUserID(command.UserID)
	if err != nil {
		return money.CategoryBudget{}, err
	}
	category, ok := money.FindCategory(money.TransactionTypeExpense, command.Category)
	if !ok {
		return money.CategoryBudget{}, errors.New("unknown expense category")
	}
	budget, err := money.NewCategoryBudget(userID, category, command.MonthlyAmountCents, false)
	if err != nil {
		return money.CategoryBudget{}, err
	}
	if err := h.budgets.Save(ctx, budget); err != nil {
		return money.CategoryBudget{}, err
	}
	return budget, nil
}

type SetDailyBudgetDisplayCommand struct {
	UserID   string
	Category string
	Enabled  bool
}

type SetDailyBudgetDisplayHandler struct {
	budgets ports.CategoryBudgetRepository
}

func NewSetDailyBudgetDisplayHandler(budgets ports.CategoryBudgetRepository) SetDailyBudgetDisplayHandler {
	return SetDailyBudgetDisplayHandler{budgets: budgets}
}

func (h SetDailyBudgetDisplayHandler) Handle(ctx context.Context, command SetDailyBudgetDisplayCommand) error {
	userID, err := money.NewUserID(command.UserID)
	if err != nil {
		return err
	}
	category, ok := money.FindCategory(money.TransactionTypeExpense, command.Category)
	if !ok {
		return errors.New("unknown expense category")
	}
	return h.budgets.SetDailyDisplay(ctx, userID, category.Code(), command.Enabled)
}

type DeleteCategoryBudgetCommand struct{ UserID, Category string }

type DeleteCategoryBudgetHandler struct {
	budgets ports.CategoryBudgetRepository
}

func NewDeleteCategoryBudgetHandler(budgets ports.CategoryBudgetRepository) DeleteCategoryBudgetHandler {
	return DeleteCategoryBudgetHandler{budgets: budgets}
}

func (h DeleteCategoryBudgetHandler) Handle(ctx context.Context, command DeleteCategoryBudgetCommand) error {
	userID, err := money.NewUserID(command.UserID)
	if err != nil {
		return err
	}
	category, ok := money.FindCategory(money.TransactionTypeExpense, command.Category)
	if !ok {
		return errors.New("unknown expense category")
	}
	return h.budgets.Delete(ctx, userID, category.Code())
}
