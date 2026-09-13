package commands

import (
	"context"
	"testing"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/application/ports"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

func TestRecordTransactionHandlerRecordsExpense(t *testing.T) {
	writer := fakeTransactionWriter{}
	handler := NewRecordTransactionHandler(writer)

	result, err := handler.Handle(context.Background(), RecordTransactionCommand{
		UserID:          "line-user-1",
		Type:            money.TransactionTypeExpense,
		AmountCents:     12500,
		Currency:        "THB",
		Category:        "food",
		Note:            "lunch",
		SourceSystem:    "line",
		SourceMessageID: "message-1",
		OccurredAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if result.Transaction.Amount().AmountCents() != 12500 {
		t.Fatalf("expected 12500 cents, got %d", result.Transaction.Amount().AmountCents())
	}
	if result.Transaction.SourceRef().MessageID() != "message-1" {
		t.Fatalf("expected source message id to be preserved")
	}
	if result.Transaction.Category().Code() != "food" {
		t.Fatalf("expected food category, got %s", result.Transaction.Category().Code())
	}
}

func TestRecordTransactionHandlerRequiresSourceMessageID(t *testing.T) {
	handler := NewRecordTransactionHandler(fakeTransactionWriter{})

	_, err := handler.Handle(context.Background(), RecordTransactionCommand{
		UserID:      "line-user-1",
		Type:        money.TransactionTypeExpense,
		AmountCents: 12500,
		Currency:    "THB",
	})
	if err == nil {
		t.Fatal("expected source message id error")
	}
}

type fakeTransactionWriter struct{}

func (fakeTransactionWriter) Save(_ context.Context, tx money.Transaction) (ports.SaveTransactionResult, error) {
	return ports.SaveTransactionResult{Transaction: tx}, nil
}
