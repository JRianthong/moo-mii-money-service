package queries

import (
	"context"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type CategoryBudgetItem struct {
	Category             string
	MonthlyBudgetCents   int64
	SpentCents           int64
	RemainingCents       int64
	AveragePerDayCents   int64
	AvailablePerDayCents int64
	ShowInDailyOverview  bool
}

type CategoryBudgetsResult struct {
	From          time.Time
	To            time.Time
	DaysInPeriod  int
	DaysRemaining int
	Items         []CategoryBudgetItem
}

type CategoryBudgetsHandler struct {
	budgets      ports.CategoryBudgetRepository
	transactions ports.TransactionReader
	settings     ports.UserSettingsRepository
}

func NewCategoryBudgetsHandler(budgets ports.CategoryBudgetRepository, transactions ports.TransactionReader, settings ports.UserSettingsRepository) CategoryBudgetsHandler {
	return CategoryBudgetsHandler{budgets: budgets, transactions: transactions, settings: settings}
}

func (h CategoryBudgetsHandler) Handle(ctx context.Context, userIDText string, now time.Time) (CategoryBudgetsResult, error) {
	userID, err := money.NewUserID(userIDText)
	if err != nil {
		return CategoryBudgetsResult{}, err
	}
	if now.IsZero() {
		now = time.Now()
	}
	cycle, err := h.settings.BillingCycle(ctx, userID)
	if err != nil {
		return CategoryBudgetsResult{}, err
	}
	from, to := cycle.Period(now)
	budgets, err := h.budgets.List(ctx, userID)
	if err != nil {
		return CategoryBudgetsResult{}, err
	}
	spent, err := h.transactions.SumExpensesByCategory(ctx, userID, from, now.Add(time.Nanosecond))
	if err != nil {
		return CategoryBudgetsResult{}, err
	}
	daysInPeriod := int(to.Sub(from).Hours() / 24)
	daysRemaining := int(to.Sub(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())).Hours() / 24)
	if daysRemaining < 1 {
		daysRemaining = 1
	}
	result := CategoryBudgetsResult{From: from, To: to, DaysInPeriod: daysInPeriod, DaysRemaining: daysRemaining}
	for _, budget := range budgets {
		used := spent[budget.Category().Code()]
		remaining := budget.MonthlyAmountCents() - used
		if remaining < 0 {
			remaining = 0
		}
		result.Items = append(result.Items, CategoryBudgetItem{
			Category: budget.Category().Name(), MonthlyBudgetCents: budget.MonthlyAmountCents(),
			SpentCents: used, RemainingCents: remaining,
			AveragePerDayCents:   budget.MonthlyAmountCents() / int64(daysInPeriod),
			AvailablePerDayCents: remaining / int64(daysRemaining),
			ShowInDailyOverview:  budget.ShowInDailyOverview(),
		})
	}
	return result, nil
}
