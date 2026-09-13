package money

import "testing"

func TestCategorizeInfersExpenseCategoryFromThaiNote(t *testing.T) {
	category, err := Categorize(TransactionTypeExpense, "", "ข้าวกลางวัน")
	if err != nil {
		t.Fatalf("categorize failed: %v", err)
	}
	if category.Code() != "food" || category.Name() != "อาหาร" {
		t.Fatalf("expected food category, got %s/%s", category.Code(), category.Name())
	}
}

func TestCategorizeInfersIncomeCategoryFromThaiNote(t *testing.T) {
	category, err := Categorize(TransactionTypeIncome, "", "เงินเดือน")
	if err != nil {
		t.Fatalf("categorize failed: %v", err)
	}
	if category.Code() != "salary" || category.Name() != "เงินเดือน" {
		t.Fatalf("expected salary category, got %s/%s", category.Code(), category.Name())
	}
}

func TestCategorizeUsesExplicitCategory(t *testing.T) {
	category, err := Categorize(TransactionTypeExpense, "transport", "grab")
	if err != nil {
		t.Fatalf("categorize failed: %v", err)
	}
	if category.Code() != "transport" {
		t.Fatalf("expected transport category, got %s", category.Code())
	}
}
