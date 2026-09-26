package gormrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CategoryBudgetModel struct {
	UserID             string `gorm:"type:text;primaryKey"`
	CategoryCode       string `gorm:"type:text;primaryKey"`
	CategoryName       string `gorm:"type:text;not null"`
	MonthlyAmountCents int64  `gorm:"not null"`
	ShowDaily          bool   `gorm:"not null;default:false"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (CategoryBudgetModel) TableName() string { return "category_budgets" }

type CategoryBudgetRepository struct{ db *gorm.DB }

func NewCategoryBudgetRepository(db *gorm.DB) CategoryBudgetRepository {
	return CategoryBudgetRepository{db: db}
}

func (r CategoryBudgetRepository) Save(ctx context.Context, budget money.CategoryBudget) error {
	model := CategoryBudgetModel{
		UserID: budget.UserID().String(), CategoryCode: budget.Category().Code(),
		CategoryName: budget.Category().Name(), MonthlyAmountCents: budget.MonthlyAmountCents(),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "category_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"category_name", "monthly_amount_cents", "updated_at"}),
	}).Create(&model).Error
}

func (r CategoryBudgetRepository) SetDailyDisplay(ctx context.Context, userID money.UserID, categoryCode string, enabled bool) error {
	result := r.db.WithContext(ctx).Model(&CategoryBudgetModel{}).
		Where("user_id = ? AND category_code = ?", userID.String(), categoryCode).
		Update("show_daily", enabled)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r CategoryBudgetRepository) List(ctx context.Context, userID money.UserID) ([]money.CategoryBudget, error) {
	var models []CategoryBudgetModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID.String()).Order("category_code").Find(&models).Error; err != nil {
		return nil, err
	}
	budgets := make([]money.CategoryBudget, 0, len(models))
	for _, model := range models {
		category, ok := money.FindCategory(money.TransactionTypeExpense, model.CategoryCode)
		if !ok {
			return nil, errors.New("stored category budget references an unknown category")
		}
		budget, err := money.NewCategoryBudget(userID, category, model.MonthlyAmountCents, model.ShowDaily)
		if err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}
	return budgets, nil
}

func (r CategoryBudgetRepository) Delete(ctx context.Context, userID money.UserID, categoryCode string) error {
	result := r.db.WithContext(ctx).Where("user_id = ? AND category_code = ?", userID.String(), categoryCode).Delete(&CategoryBudgetModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

var _ ports.CategoryBudgetRepository = CategoryBudgetRepository{}
