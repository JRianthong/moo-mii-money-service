package line

import (
	"testing"
	"time"

	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

func TestTextParserParsesThaiExpense(t *testing.T) {
	parser := NewTextParser("THB", time.UTC)
	now := time.Date(2026, 9, 13, 9, 0, 0, 0, time.UTC)

	intent, err := parser.Parse("line-user-1", "message-1", "จ่าย 120.50 ข้าวกลางวัน", now)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if intent.Type != ParsedIntentRecord {
		t.Fatalf("expected record intent, got %s", intent.Type)
	}
	if intent.Command.Type != money.TransactionTypeExpense {
		t.Fatalf("expected expense, got %s", intent.Command.Type)
	}
	if intent.Command.AmountCents != 12050 {
		t.Fatalf("expected 12050 cents, got %d", intent.Command.AmountCents)
	}
	if intent.Command.Category != "" {
		t.Fatalf("expected categorization to happen in application layer, got %s", intent.Command.Category)
	}
}

func TestTextParserParsesSignedIncome(t *testing.T) {
	parser := NewTextParser("THB", time.UTC)

	intent, err := parser.Parse("line-user-1", "message-1", "+3,000 เงินเดือน", time.Now())
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if intent.Command.Type != money.TransactionTypeIncome {
		t.Fatalf("expected income, got %s", intent.Command.Type)
	}
	if intent.Command.AmountCents != 300000 {
		t.Fatalf("expected 300000 cents, got %d", intent.Command.AmountCents)
	}
}

func TestTextParserExtractsExplicitCategory(t *testing.T) {
	parser := NewTextParser("THB", time.UTC)

	intent, err := parser.Parse("line-user-1", "message-1", "จ่าย 120 #เดินทาง grab", time.Now())
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if intent.Command.Category != "เดินทาง" {
		t.Fatalf("expected explicit category, got %s", intent.Command.Category)
	}
	if intent.Command.Note != "grab" {
		t.Fatalf("expected category token removed from note, got %s", intent.Command.Note)
	}
}

func TestTextParserParsesSummaryIntent(t *testing.T) {
	parser := NewTextParser("THB", time.UTC)

	intent, err := parser.Parse("line-user-1", "message-1", "สรุป", time.Now())
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if intent.Type != ParsedIntentSummary {
		t.Fatalf("expected summary intent, got %s", intent.Type)
	}
}

func TestTextParserParsesCategoriesIntent(t *testing.T) {
	parser := NewTextParser("THB", time.UTC)

	intent, err := parser.Parse("line-user-1", "message-1", "หมวดหมู่", time.Now())
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if intent.Type != ParsedIntentCategories {
		t.Fatalf("expected categories intent, got %s", intent.Type)
	}
}
