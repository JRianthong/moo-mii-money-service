package money

import "errors"

type CategoryBudget struct {
	userID              UserID
	category            Category
	monthlyAmountCents  int64
	showInDailyOverview bool
}

func NewCategoryBudget(userID UserID, category Category, monthlyAmountCents int64, showInDailyOverview bool) (CategoryBudget, error) {
	if userID.String() == "" {
		return CategoryBudget{}, errors.New("user id is required")
	}
	if err := category.ValidateFor(TransactionTypeExpense); err != nil {
		return CategoryBudget{}, err
	}
	if monthlyAmountCents <= 0 {
		return CategoryBudget{}, errors.New("monthly budget must be greater than zero")
	}
	return CategoryBudget{
		userID: userID, category: category, monthlyAmountCents: monthlyAmountCents,
		showInDailyOverview: showInDailyOverview,
	}, nil
}

func (b CategoryBudget) UserID() UserID            { return b.userID }
func (b CategoryBudget) Category() Category        { return b.category }
func (b CategoryBudget) MonthlyAmountCents() int64 { return b.monthlyAmountCents }
func (b CategoryBudget) ShowInDailyOverview() bool { return b.showInDailyOverview }
