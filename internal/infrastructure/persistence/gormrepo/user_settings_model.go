package gormrepo

import (
	"context"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserSettingsModel struct {
	UserID          string `gorm:"type:text;primaryKey"`
	BillingCycleDay int    `gorm:"not null;default:1"`
	UpdatedAt       time.Time
}

func (UserSettingsModel) TableName() string { return "user_settings" }

type UserSettingsRepository struct {
	db *gorm.DB
}

func NewUserSettingsRepository(db *gorm.DB) UserSettingsRepository {
	return UserSettingsRepository{db: db}
}

func (r UserSettingsRepository) BillingCycle(ctx context.Context, userID money.UserID) (money.BillingCycle, error) {
	var model UserSettingsModel
	err := r.db.WithContext(ctx).First(&model, "user_id = ?", userID.String()).Error
	if err == gorm.ErrRecordNotFound {
		return money.DefaultBillingCycle(), nil
	}
	if err != nil {
		return money.BillingCycle{}, err
	}
	return money.NewBillingCycle(model.BillingCycleDay)
}

func (r UserSettingsRepository) SetBillingCycle(ctx context.Context, userID money.UserID, cycle money.BillingCycle) error {
	model := UserSettingsModel{UserID: userID.String(), BillingCycleDay: cycle.StartDay()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"billing_cycle_day", "updated_at"}),
	}).Create(&model).Error
}
