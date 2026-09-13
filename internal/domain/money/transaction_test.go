package money

import (
	"testing"
	"time"
)

func TestNewTransactionDefaultsCurrencyAndValidates(t *testing.T) {
	tx, err := newTestExpense(12000, "")
	if err != nil {
		t.Fatalf("expected valid transaction: %v", err)
	}
	if tx.Amount().Currency() != "THB" {
		t.Fatalf("expected default currency THB, got %s", tx.Amount().Currency())
	}
}

func TestNewTransactionRejectsInvalidAmount(t *testing.T) {
	_, err := newTestExpense(0, "THB")
	if err == nil {
		t.Fatal("expected invalid amount error")
	}
}

func newTestExpense(amountCents int64, currency string) (Transaction, error) {
	userID, err := NewUserID("user-1")
	if err != nil {
		return Transaction{}, err
	}
	amount, err := NewMoney(amountCents, currency)
	if err != nil {
		return Transaction{}, err
	}
	sourceRef, err := NewSourceRef("line", "message-1")
	if err != nil {
		return Transaction{}, err
	}
	return RecordExpense(RecordTransactionInput{
		UserID:     userID,
		Amount:     amount,
		Category:   mustTestCategory(TransactionTypeExpense, "food", "อาหาร"),
		Note:       NewNote("lunch"),
		SourceRef:  sourceRef,
		OccurredAt: NewOccurredAt(time.Now()),
	})
}

func TestTransactionRejectsCategoryTypeMismatch(t *testing.T) {
	userID, err := NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}
	amount, err := NewMoney(12000, "THB")
	if err != nil {
		t.Fatal(err)
	}
	sourceRef, err := NewSourceRef("line", "message-1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = RecordExpense(RecordTransactionInput{
		UserID:     userID,
		Amount:     amount,
		Category:   mustTestCategory(TransactionTypeIncome, "salary", "เงินเดือน"),
		Note:       NewNote("wrong category"),
		SourceRef:  sourceRef,
		OccurredAt: NewOccurredAt(time.Now()),
	})
	if err == nil {
		t.Fatal("expected category type mismatch error")
	}
}

func mustTestCategory(txType TransactionType, code, name string) Category {
	category, err := NewCategory(txType, code, name)
	if err != nil {
		panic(err)
	}
	return category
}
