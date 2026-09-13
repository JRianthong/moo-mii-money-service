package gormrepo

import (
	"context"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return TransactionRepository{db: db}
}

func (r TransactionRepository) Save(ctx context.Context, tx money.Transaction) (ports.SaveTransactionResult, error) {
	model := transactionModelFromDomain(tx)
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source_system"}, {Name: "source_message_id"}},
		DoNothing: true,
	}).Create(&model)
	if result.Error != nil {
		return ports.SaveTransactionResult{}, result.Error
	}
	if result.RowsAffected == 0 {
		existing, err := r.findBySourceRef(ctx, tx.SourceRef())
		if err != nil {
			return ports.SaveTransactionResult{}, err
		}
		return ports.SaveTransactionResult{Transaction: existing, AlreadyRecorded: true}, nil
	}
	saved, err := model.toDomain()
	if err != nil {
		return ports.SaveTransactionResult{}, err
	}
	return ports.SaveTransactionResult{Transaction: saved}, nil
}

func (r TransactionRepository) SumByType(ctx context.Context, userID money.UserID, from, to time.Time) (int64, int64, error) {
	type row struct {
		Type  string
		Total int64
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&TransactionModel{}).
		Select("type, COALESCE(SUM(amount_cents), 0) AS total").
		Where("user_id = ? AND occurred_at >= ? AND occurred_at < ?", userID.String(), from, to).
		Group("type").
		Scan(&rows).
		Error
	if err != nil {
		return 0, 0, err
	}

	var income, expense int64
	for _, row := range rows {
		switch money.TransactionType(row.Type) {
		case money.TransactionTypeIncome:
			income = row.Total
		case money.TransactionTypeExpense:
			expense = row.Total
		}
	}
	return income, expense, nil
}

func (r TransactionRepository) ListRecent(ctx context.Context, userID money.UserID, limit int) ([]money.Transaction, error) {
	var models []TransactionModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID.String()).
		Order("occurred_at DESC, created_at DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}

	transactions := make([]money.Transaction, 0, len(models))
	for _, model := range models {
		tx, err := model.toDomain()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func (r TransactionRepository) findBySourceRef(ctx context.Context, sourceRef money.SourceRef) (money.Transaction, error) {
	var model TransactionModel
	if err := r.db.WithContext(ctx).
		Where("source_system = ? AND source_message_id = ?", sourceRef.System(), sourceRef.MessageID()).
		First(&model).Error; err != nil {
		return money.Transaction{}, err
	}
	return model.toDomain()
}
