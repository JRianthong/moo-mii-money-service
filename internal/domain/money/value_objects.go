package money

import (
	"errors"
	"strings"
	"time"
)

type TransactionID string

func (id TransactionID) String() string {
	return string(id)
}

type UserID string

func NewUserID(value string) (UserID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("user id is required")
	}
	return UserID(value), nil
}

func (id UserID) String() string {
	return string(id)
}

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

func (t TransactionType) IsValid() bool {
	return t == TransactionTypeIncome || t == TransactionTypeExpense
}

type Money struct {
	amountCents int64
	currency    string
}

func NewMoney(amountCents int64, currency string) (Money, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "THB"
	}
	if amountCents <= 0 {
		return Money{}, errors.New("amount must be greater than zero")
	}
	return Money{amountCents: amountCents, currency: currency}, nil
}

func (m Money) AmountCents() int64 {
	return m.amountCents
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Float64() float64 {
	return float64(m.amountCents) / 100
}

type Note string

func NewNote(value string) Note {
	return Note(strings.TrimSpace(value))
}

func (n Note) String() string {
	return string(n)
}

type SourceRef struct {
	system    string
	messageID string
}

func NewSourceRef(system, messageID string) (SourceRef, error) {
	system = strings.TrimSpace(system)
	messageID = strings.TrimSpace(messageID)
	if system == "" {
		return SourceRef{}, errors.New("source system is required")
	}
	if messageID == "" {
		return SourceRef{}, errors.New("source message id is required")
	}
	return SourceRef{system: system, messageID: messageID}, nil
}

func (s SourceRef) System() string {
	return s.system
}

func (s SourceRef) MessageID() string {
	return s.messageID
}

type OccurredAt struct {
	value time.Time
}

func NewOccurredAt(value time.Time) OccurredAt {
	if value.IsZero() {
		value = time.Now()
	}
	return OccurredAt{value: value}
}

func (o OccurredAt) Time() time.Time {
	return o.value
}
