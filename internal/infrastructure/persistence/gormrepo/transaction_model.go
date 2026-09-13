package gormrepo

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
	"gorm.io/gorm"
)

type TransactionModel struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	UserID          string `gorm:"type:text;not null;index:idx_transactions_user_occurred,priority:1"`
	Type            string `gorm:"type:text;not null;index"`
	AmountCents     int64  `gorm:"not null"`
	Currency        string `gorm:"type:text;not null"`
	CategoryCode    string `gorm:"type:text;not null"`
	CategoryName    string `gorm:"type:text;not null"`
	Note            string `gorm:"type:text"`
	SourceSystem    string `gorm:"type:text;not null;uniqueIndex:idx_transactions_source_message"`
	SourceMessageID string `gorm:"type:text;not null;uniqueIndex:idx_transactions_source_message"`
	OccurredAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (TransactionModel) TableName() string {
	return "transactions"
}

func (m *TransactionModel) BeforeCreate(_ *gorm.DB) error {
	if m.ID != "" {
		return nil
	}
	id, err := newUUID()
	if err != nil {
		return err
	}
	m.ID = id
	return nil
}

func transactionModelFromDomain(tx money.Transaction) TransactionModel {
	return TransactionModel{
		ID:              tx.ID().String(),
		UserID:          tx.UserID().String(),
		Type:            string(tx.Type()),
		AmountCents:     tx.Amount().AmountCents(),
		Currency:        tx.Amount().Currency(),
		CategoryCode:    tx.Category().Code(),
		CategoryName:    tx.Category().Name(),
		Note:            tx.Note().String(),
		SourceSystem:    tx.SourceRef().System(),
		SourceMessageID: tx.SourceRef().MessageID(),
		OccurredAt:      tx.OccurredAt(),
		CreatedAt:       tx.CreatedAt(),
	}
}

func (m TransactionModel) toDomain() (money.Transaction, error) {
	userID, err := money.NewUserID(m.UserID)
	if err != nil {
		return money.Transaction{}, err
	}
	amount, err := money.NewMoney(m.AmountCents, m.Currency)
	if err != nil {
		return money.Transaction{}, err
	}
	sourceRef, err := money.NewSourceRef(m.SourceSystem, m.SourceMessageID)
	if err != nil {
		return money.Transaction{}, err
	}
	category, err := money.NewCategory(money.TransactionType(m.Type), m.CategoryCode, m.CategoryName)
	if err != nil {
		return money.Transaction{}, err
	}
	return money.RehydrateTransaction(
		money.TransactionID(m.ID),
		money.RecordTransactionInput{
			UserID:     userID,
			Type:       money.TransactionType(m.Type),
			Amount:     amount,
			Category:   category,
			Note:       money.NewNote(m.Note),
			SourceRef:  sourceRef,
			OccurredAt: money.NewOccurredAt(m.OccurredAt),
		},
		m.CreatedAt,
	)
}

func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
