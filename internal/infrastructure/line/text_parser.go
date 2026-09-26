package line

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jatuporn/moo-mii-money-service/internal/application/commands"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
)

type ParsedIntentType string

const (
	ParsedIntentRecord                ParsedIntentType = "record"
	ParsedIntentSummary               ParsedIntentType = "summary"
	ParsedIntentRecent                ParsedIntentType = "recent"
	ParsedIntentCategories            ParsedIntentType = "categories"
	ParsedIntentBillingCycle          ParsedIntentType = "billing_cycle"
	ParsedIntentSetBillingCycle       ParsedIntentType = "set_billing_cycle"
	ParsedIntentCategoryBudgets       ParsedIntentType = "category_budgets"
	ParsedIntentDailyBudgets          ParsedIntentType = "daily_budgets"
	ParsedIntentSetCategoryBudget     ParsedIntentType = "set_category_budget"
	ParsedIntentSetDailyBudgetDisplay ParsedIntentType = "set_daily_budget_display"
	ParsedIntentDeleteCategoryBudget  ParsedIntentType = "delete_category_budget"
	ParsedIntentHelp                  ParsedIntentType = "help"
	ParsedIntentUnknown               ParsedIntentType = "unknown"
)

type ParsedIntent struct {
	Type                ParsedIntentType
	Command             commands.RecordTransactionCommand
	BillingCycleDay     int
	BudgetCategory      string
	BudgetAmountCents   int64
	DailyDisplayEnabled bool
}

type TextParser struct {
	defaultCurrency string
	location        *time.Location
}

func NewTextParser(defaultCurrency string, location *time.Location) TextParser {
	if location == nil {
		location = time.Local
	}
	return TextParser{defaultCurrency: defaultCurrency, location: location}
}

func (p TextParser) Parse(userID, sourceMessageID, text string, now time.Time) (ParsedIntent, error) {
	normalized := strings.TrimSpace(text)
	lower := strings.ToLower(normalized)
	if normalized == "" {
		return ParsedIntent{Type: ParsedIntentUnknown}, nil
	}
	if lower == "help" || lower == "วิธีใช้" || lower == "ช่วยเหลือ" {
		return ParsedIntent{Type: ParsedIntentHelp}, nil
	}
	if lower == "summary" || lower == "สรุป" || lower == "สรุปเดือนนี้" {
		return ParsedIntent{Type: ParsedIntentSummary}, nil
	}
	if lower == "recent" || lower == "ล่าสุด" || lower == "รายการล่าสุด" {
		return ParsedIntent{Type: ParsedIntentRecent}, nil
	}
	if lower == "categories" || lower == "category" || lower == "หมวดหมู่" || lower == "หมวด" {
		return ParsedIntent{Type: ParsedIntentCategories}, nil
	}
	if lower == "รอบ" || lower == "ตั้งรอบ" {
		return ParsedIntent{Type: ParsedIntentBillingCycle}, nil
	}
	if lower == "งบ" || lower == "งบเดือนนี้" {
		return ParsedIntent{Type: ParsedIntentCategoryBudgets}, nil
	}
	if lower == "งบรายวัน" || lower == "งบวันนี้" {
		return ParsedIntent{Type: ParsedIntentDailyBudgets}, nil
	}
	if strings.HasPrefix(lower, "ตั้งงบ ") {
		fields := strings.Fields(normalized)
		if len(fields) < 3 {
			return ParsedIntent{Type: ParsedIntentSetCategoryBudget}, errors.New("use: ตั้งงบ <หมวดหมู่> <จำนวนเงิน>")
		}
		amount, err := parseAmountCents(fields[len(fields)-1])
		if err != nil {
			return ParsedIntent{Type: ParsedIntentSetCategoryBudget}, err
		}
		return ParsedIntent{Type: ParsedIntentSetCategoryBudget, BudgetCategory: strings.Join(fields[1:len(fields)-1], " "), BudgetAmountCents: amount}, nil
	}
	if strings.HasPrefix(lower, "แสดงงบรายวัน ") || strings.HasPrefix(lower, "ซ่อนงบรายวัน ") {
		fields := strings.Fields(normalized)
		if len(fields) != 2 {
			prefix := "แสดงงบรายวัน "
			if strings.HasPrefix(lower, "ซ่อนงบรายวัน ") {
				prefix = "ซ่อนงบรายวัน "
			}
			category := strings.TrimSpace(normalized[len(prefix):])
			if category == "" {
				return ParsedIntent{Type: ParsedIntentSetDailyBudgetDisplay}, errors.New("specify one expense category")
			}
			enabled := strings.HasPrefix(lower, "แสดงงบรายวัน ")
			return ParsedIntent{Type: ParsedIntentSetDailyBudgetDisplay, BudgetCategory: category, DailyDisplayEnabled: enabled}, nil
		}
		enabled := strings.HasPrefix(lower, "แสดงงบรายวัน ")
		return ParsedIntent{Type: ParsedIntentSetDailyBudgetDisplay, BudgetCategory: fields[1], DailyDisplayEnabled: enabled}, nil
	}
	if strings.HasPrefix(lower, "ลบงบ ") {
		category := strings.TrimSpace(normalized[len("ลบงบ "):])
		if category == "" {
			return ParsedIntent{Type: ParsedIntentDeleteCategoryBudget}, errors.New("specify one expense category")
		}
		return ParsedIntent{Type: ParsedIntentDeleteCategoryBudget, BudgetCategory: category}, nil
	}
	if strings.HasPrefix(lower, "ตั้งรอบ ") || strings.HasPrefix(lower, "ตัดรอบ ") {
		fields := strings.Fields(normalized)
		if len(fields) != 2 {
			return ParsedIntent{Type: ParsedIntentSetBillingCycle}, errors.New("billing cycle day is required")
		}
		day, err := strconv.Atoi(fields[1])
		if err != nil || day < 1 || day > 31 {
			return ParsedIntent{Type: ParsedIntentSetBillingCycle}, errors.New("billing cycle day must be between 1 and 31")
		}
		return ParsedIntent{Type: ParsedIntentSetBillingCycle, BillingCycleDay: day}, nil
	}

	txType, rest, ok := detectType(normalized)
	if !ok {
		return ParsedIntent{Type: ParsedIntentUnknown}, nil
	}

	amountToken, note, ok := splitAmount(rest)
	if !ok {
		return ParsedIntent{Type: ParsedIntentRecord}, errors.New("amount is required")
	}
	amountCents, err := parseAmountCents(amountToken)
	if err != nil {
		return ParsedIntent{Type: ParsedIntentRecord}, err
	}
	category, note := extractExplicitCategory(note)

	return ParsedIntent{
		Type: ParsedIntentRecord,
		Command: commands.RecordTransactionCommand{
			UserID:          userID,
			Type:            txType,
			AmountCents:     amountCents,
			Currency:        p.defaultCurrency,
			Category:        category,
			Note:            note,
			SourceSystem:    "line",
			SourceMessageID: sourceMessageID,
			OccurredAt:      now.In(p.location),
		},
	}, nil
}

func detectType(text string) (money.TransactionType, string, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "+"):
		return money.TransactionTypeIncome, strings.TrimSpace(trimmed[1:]), true
	case strings.HasPrefix(lower, "-"):
		return money.TransactionTypeExpense, strings.TrimSpace(trimmed[1:]), true
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", "", false
	}
	first := strings.ToLower(fields[0])
	rest := strings.TrimSpace(strings.TrimPrefix(trimmed, fields[0]))
	switch first {
	case "income", "in", "รับ", "รายรับ", "ได้เงิน":
		return money.TransactionTypeIncome, rest, true
	case "expense", "out", "จ่าย", "รายจ่าย", "ซื้อ":
		return money.TransactionTypeExpense, rest, true
	default:
		return "", "", false
	}
}

func splitAmount(text string) (string, string, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", "", false
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "", "", false
	}
	amount := strings.Trim(fields[0], "฿บาท")
	note := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	return amount, note, true
}

func parseAmountCents(text string) (int64, error) {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '.' {
			return r
		}
		if r == ',' || unicode.IsSpace(r) {
			return -1
		}
		return -1
	}, text)
	amount, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, errors.New("amount must be a number")
	}
	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}
	return int64(math.Round(amount * 100)), nil
}

func extractExplicitCategory(note string) (string, string) {
	fields := strings.Fields(strings.TrimSpace(note))
	for i, field := range fields {
		if strings.HasPrefix(field, "#") && len(field) > 1 {
			category := strings.TrimPrefix(field, "#")
			fields = append(fields[:i], fields[i+1:]...)
			return category, strings.Join(fields, " ")
		}
		if strings.HasPrefix(field, "หมวด:") && len(field) > len("หมวด:") {
			category := strings.TrimPrefix(field, "หมวด:")
			fields = append(fields[:i], fields[i+1:]...)
			return category, strings.Join(fields, " ")
		}
	}
	return "", note
}
