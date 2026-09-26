package http

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jatuporn/moo-mii-money-service/internal/application/commands"
	"github.com/jatuporn/moo-mii-money-service/internal/application/queries"
	"github.com/jatuporn/moo-mii-money-service/internal/domain/money"
	"github.com/jatuporn/moo-mii-money-service/internal/infrastructure/line"
)

type Dependencies struct {
	LineClient             line.Client
	TextParser             line.TextParser
	RecordTransaction      commands.RecordTransactionHandler
	SetBillingCycle        commands.SetBillingCycleHandler
	SetCategoryBudget      commands.SetCategoryBudgetHandler
	SetDailyBudgetDisplay  commands.SetDailyBudgetDisplayHandler
	DeleteCategoryBudget   commands.DeleteCategoryBudgetHandler
	GetMonthlySummary      queries.MonthlySummaryHandler
	GetBillingCycle        queries.GetBillingCycleHandler
	GetCategoryBudgets     queries.CategoryBudgetsHandler
	ListRecentTransactions queries.ListRecentTransactionsHandler
}

func RegisterRoutes(app *fiber.App, deps Dependencies) {
	healthz := func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	}
	app.Get("/healthz", healthz)
	app.Get("/api/healthz", healthz)

	lineWebhook := func(c *fiber.Ctx) error {
		if !deps.LineClient.ValidateSignature(c.Body(), c.Get("X-Line-Signature")) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid signature")
		}

		var req line.WebhookRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		for _, event := range req.Events {
			if event.Type != "message" || event.Message == nil || event.Message.Type != "text" {
				continue
			}
			if err := handleLineTextEvent(c.UserContext(), deps, event, time.Now()); err != nil {
				return err
			}
		}
		return c.SendStatus(fiber.StatusOK)
	}
	app.Post("/webhooks/line", lineWebhook)
	app.Post("/api/webhooks/line", lineWebhook)
}

func handleLineTextEvent(ctx context.Context, deps Dependencies, event line.WebhookEvent, now time.Time) error {
	userID := event.Source.UserID
	if userID == "" {
		userID = event.Source.GroupID
	}
	if userID == "" {
		userID = event.Source.RoomID
	}

	intent, err := deps.TextParser.Parse(userID, event.Message.ID, event.Message.Text, now)
	if err != nil {
		if strings.Contains(strings.ToLower(event.Message.Text), "รอบ") {
			return deps.LineClient.ReplyText(ctx, event.ReplyToken, "ตั้งรอบได้ตั้งแต่วันที่ 1 ถึง 31 เช่น ตั้งรอบ 28")
		}
		if strings.Contains(strings.ToLower(event.Message.Text), "งบ") {
			return deps.LineClient.ReplyText(ctx, event.ReplyToken, "ตัวอย่าง: ตั้งงบ อาหาร 9000 หรือ แสดงงบรายวัน อาหาร")
		}
		return deps.LineClient.ReplyText(ctx, event.ReplyToken, "อ่านยอดไม่สำเร็จ ลองพิมพ์เช่น: จ่าย 120 ข้าวกลางวัน")
	}

	var reply string
	switch intent.Type {
	case line.ParsedIntentRecord:
		result, err := deps.RecordTransaction.Handle(ctx, intent.Command)
		if err != nil {
			return err
		}
		var budget *queries.CategoryBudgetItem
		if result.Transaction.Type() == money.TransactionTypeExpense {
			if budgets, err := deps.GetCategoryBudgets.Handle(ctx, userID, now); err == nil {
				for i := range budgets.Items {
					if budgets.Items[i].Category == result.Transaction.Category().Name() {
						budget = &budgets.Items[i]
						break
					}
				}
			}
		}
		altText := "บันทึกรายการสำเร็จ"
		if result.AlreadyRecorded {
			altText = "รายการนี้ถูกบันทึกแล้ว"
		}
		return deps.LineClient.ReplyFlex(ctx, event.ReplyToken, altText, formatRecordedFlex(result.Transaction, result.AlreadyRecorded, budget))
	case line.ParsedIntentSummary:
		summary, err := deps.GetMonthlySummary.Handle(ctx, queries.MonthlySummaryQuery{UserID: userID, Now: now})
		if err != nil {
			return err
		}
		reply = formatSummary(summary)
	case line.ParsedIntentRecent:
		transactions, err := deps.ListRecentTransactions.Handle(ctx, queries.ListRecentTransactionsQuery{UserID: userID, Limit: 5})
		if err != nil {
			return err
		}
		reply = formatRecent(transactions)
	case line.ParsedIntentCategories:
		reply = formatCategories()
	case line.ParsedIntentBillingCycle:
		cycle, err := deps.GetBillingCycle.Handle(ctx, userID)
		if err != nil {
			return err
		}
		reply = fmt.Sprintf("ตอนนี้ตัดรอบทุกวันที่ %d\nเปลี่ยนได้ด้วยคำสั่ง: ตั้งรอบ 28", cycle.StartDay())
	case line.ParsedIntentSetBillingCycle:
		cycle, err := deps.SetBillingCycle.Handle(ctx, commands.SetBillingCycleCommand{UserID: userID, StartDay: intent.BillingCycleDay})
		if err != nil {
			return err
		}
		reply = fmt.Sprintf("ตั้งรอบเรียบร้อย ตัดรอบทุกวันที่ %d\nพิมพ์ สรุป เพื่อดูยอดตามรอบใหม่", cycle.StartDay())
	case line.ParsedIntentCategoryBudgets, line.ParsedIntentDailyBudgets:
		budgets, err := deps.GetCategoryBudgets.Handle(ctx, userID, now)
		if err != nil {
			return err
		}
		reply = formatCategoryBudgets(budgets, intent.Type == line.ParsedIntentDailyBudgets)
	case line.ParsedIntentSetCategoryBudget:
		budget, err := deps.SetCategoryBudget.Handle(ctx, commands.SetCategoryBudgetCommand{
			UserID: userID, Category: intent.BudgetCategory, MonthlyAmountCents: intent.BudgetAmountCents,
		})
		if err != nil {
			return deps.LineClient.ReplyText(ctx, event.ReplyToken, "ตั้งงบไม่สำเร็จ ตรวจชื่อหมวดและจำนวนเงิน เช่น ตั้งงบ อาหาร 9000")
		}
		reply = fmt.Sprintf("ตั้งงบหมวด%s เดือนละ %.2f บาทแล้ว", budget.Category().Name(), centsToFloat(budget.MonthlyAmountCents()))
	case line.ParsedIntentSetDailyBudgetDisplay:
		err := deps.SetDailyBudgetDisplay.Handle(ctx, commands.SetDailyBudgetDisplayCommand{
			UserID: userID, Category: intent.BudgetCategory, Enabled: intent.DailyDisplayEnabled,
		})
		if err != nil {
			return deps.LineClient.ReplyText(ctx, event.ReplyToken, "เปลี่ยนการแสดงงบรายวันไม่สำเร็จ ตั้งงบหมวดนั้นก่อน เช่น ตั้งงบ อาหาร 9000")
		}
		status := "แสดง"
		if !intent.DailyDisplayEnabled {
			status = "ซ่อน"
		}
		reply = fmt.Sprintf("%sหมวด%sในงบรายวันแล้ว", status, intent.BudgetCategory)
	case line.ParsedIntentDeleteCategoryBudget:
		if err := deps.DeleteCategoryBudget.Handle(ctx, commands.DeleteCategoryBudgetCommand{UserID: userID, Category: intent.BudgetCategory}); err != nil {
			return deps.LineClient.ReplyText(ctx, event.ReplyToken, "ลบงบไม่สำเร็จ ตรวจชื่อหมวดและลองอีกครั้ง")
		}
		reply = "ลบงบหมวด" + intent.BudgetCategory + "แล้ว"
	case line.ParsedIntentHelp:
		reply = helpText()
	default:
		reply = helpText()
	}

	return deps.LineClient.ReplyText(ctx, event.ReplyToken, reply)
}

func formatCategories() string {
	expense := formatCategoryLine("รายจ่าย", money.AvailableCategories(money.TransactionTypeExpense))
	income := formatCategoryLine("รายรับ", money.AvailableCategories(money.TransactionTypeIncome))
	return expense + "\n" + income
}

func formatCategoryLine(label string, categories []money.Category) string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		names = append(names, "#"+category.Name())
	}
	return label + ": " + strings.Join(names, " ")
}

func formatRecordedFlex(tx money.Transaction, alreadyRecorded bool, budget *queries.CategoryBudgetItem) map[string]any {
	status := "✓ บันทึกรายการสำเร็จ"
	subtitle := "รายการของคุณถูกบันทึกแล้ว"
	if alreadyRecorded {
		status = "รายการนี้บันทึกแล้ว"
		subtitle = "รายการซ้ำจึงไม่ได้บันทึกเพิ่ม"
	}
	typeLabel := "รายจ่าย"
	amountColor := "#D65F87"
	amountPrefix := "− "
	if tx.Type() == money.TransactionTypeIncome {
		typeLabel = "รายรับ"
		amountColor = "#299B79"
		amountPrefix = "+ "
	}
	note := emptyFallback(tx.Note().String(), tx.Category().Name())
	detailContents := []any{
		flexText(typeLabel, "sm", "#B17A91", "regular"),
		flexText(amountPrefix+formatBaht(tx.Amount().AmountCents()), "xxl", amountColor, "bold"),
		map[string]any{"type": "separator", "color": "#F5E1E8", "margin": "md"},
		flexKeyValue("ชื่อรายการ", note),
		flexKeyValue("หมวดหมู่", tx.Category().Name()),
	}
	bodyContents := []any{
		map[string]any{"type": "box", "layout": "vertical", "backgroundColor": "#FFFFFF", "cornerRadius": "14px", "paddingAll": "18px", "contents": detailContents},
	}
	if budget != nil {
		budgetRows := []any{
			flexKeyValue("งบรายวันเฉลี่ย", formatBaht(budget.AveragePerDayCents)+" / วัน"),
			flexKeyValue("งบรายเดือน", formatBaht(budget.SpentCents)+" / "+formatBaht(budget.MonthlyBudgetCents)),
			flexKeyValue("เหลือเฉลี่ยต่อวันที่เหลือ", formatBaht(budget.AvailablePerDayCents)+" / วัน"),
		}
		budgetContents := []any{flexText("สรุปงบประมาณ · "+budget.Category, "sm", "#7A3E56", "bold")}
		budgetContents = append(budgetContents, budgetRows...)
		bodyContents = append(bodyContents, map[string]any{
			"type": "box", "layout": "vertical", "backgroundColor": "#FFF0F5", "cornerRadius": "14px",
			"paddingAll": "16px", "spacing": "sm", "contents": budgetContents,
		})
	}
	return map[string]any{
		"type": "bubble", "size": "mega",
		"header": map[string]any{
			"type": "box", "layout": "horizontal", "backgroundColor": "#F8BBD0", "paddingAll": "18px",
			"contents": []any{map[string]any{"type": "box", "layout": "vertical", "flex": 1, "contents": []any{
				flexText(status, "lg", "#7A3E56", "bold"),
				flexText(subtitle, "xs", "#9B6078", "regular"),
			}}},
		},
		"body": map[string]any{
			"type": "box", "layout": "vertical", "backgroundColor": "#FFF8FB", "paddingAll": "18px", "spacing": "md",
			"contents": bodyContents,
		},
	}
}

func flexText(text, size, color, weight string) map[string]any {
	return map[string]any{"type": "text", "text": text, "size": size, "color": color, "weight": weight, "wrap": true}
}

func flexKeyValue(key, value string) map[string]any {
	return map[string]any{
		"type": "box", "layout": "horizontal", "margin": "sm",
		"contents": []any{
			map[string]any{"type": "text", "text": key, "size": "xs", "color": "#999999", "flex": 1, "wrap": true},
			map[string]any{"type": "text", "text": value, "size": "xs", "weight": "bold", "color": "#6F4A5A", "align": "end", "flex": 2, "wrap": true},
		},
	}
}

func formatBaht(cents int64) string {
	whole := strconv.FormatInt(cents/100, 10)
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}
	return fmt.Sprintf("฿%s.%02d", whole, cents%100)
}

func formatSummary(summary queries.MonthlySummaryResult) string {
	return fmt.Sprintf(
		"สรุปรอบ %s - %s\nรายรับ %.2f\nรายจ่าย %.2f\nคงเหลือ %.2f",
		summary.From.Format("02/01/2006"),
		summary.To.AddDate(0, 0, -1).Format("02/01/2006"),
		centsToFloat(summary.IncomeCents),
		centsToFloat(summary.ExpenseCents),
		centsToFloat(summary.BalanceCents),
	)
}

func formatCategoryBudgets(result queries.CategoryBudgetsResult, dailyOnly bool) string {
	lines := []string{"งบรายวัน"}
	if !dailyOnly {
		lines[0] = "งบตามรอบบัญชี"
	}
	count := 0
	for _, item := range result.Items {
		if dailyOnly && !item.ShowInDailyOverview {
			continue
		}
		count++
		if dailyOnly {
			lines = append(lines, fmt.Sprintf("%s เฉลี่ย %.2f/วัน | ใช้แล้ว %.2f | เหลือ %.2f (เฉลี่ย %.2f/วันที่เหลือ)",
				item.Category, centsToFloat(item.AveragePerDayCents), centsToFloat(item.SpentCents),
				centsToFloat(item.RemainingCents), centsToFloat(item.AvailablePerDayCents)))
		} else {
			lines = append(lines, fmt.Sprintf("%s งบ %.2f | ใช้แล้ว %.2f | คงเหลือ %.2f%s",
				item.Category, centsToFloat(item.MonthlyBudgetCents), centsToFloat(item.SpentCents),
				centsToFloat(item.RemainingCents), dailyDisplaySuffix(item.ShowInDailyOverview)))
		}
	}
	if count == 0 {
		if dailyOnly {
			return "ยังไม่มีหมวดงบรายวันที่เลือก\nใช้ แสดงงบรายวัน อาหาร เพื่อเพิ่มหมวด"
		}
		return "ยังไม่มีงบรายหมวด\nตั้งงบ อาหาร 9000"
	}
	return strings.Join(lines, "\n")
}

func dailyDisplaySuffix(enabled bool) string {
	if enabled {
		return " | แสดงรายวัน"
	}
	return ""
}

func formatRecent(transactions []money.Transaction) string {
	if len(transactions) == 0 {
		return "ยังไม่มีรายการล่าสุด"
	}
	lines := []string{"รายการล่าสุด"}
	for _, tx := range transactions {
		sign := "-"
		if tx.Type() == money.TransactionTypeIncome {
			sign = "+"
		}
		lines = append(lines, fmt.Sprintf("%s %.2f %s", sign, tx.Amount().Float64(), emptyFallback(tx.Note().String(), tx.Category().String())))
	}
	return strings.Join(lines, "\n")
}

func helpText() string {
	return strings.Join([]string{
		"พิมพ์บันทึกได้แบบนี้",
		"จ่าย 120 ข้าวกลางวัน",
		"รับ 2500 ฟรีแลนซ์",
		"จ่าย 120 #เดินทาง grab",
		"-45 กาแฟ",
		"+3000 เงินเดือน",
		"สรุป",
		"ล่าสุด",
		"หมวดหมู่",
		"รอบ",
		"ตั้งรอบ 28",
		"ตั้งงบ อาหาร 9000",
		"งบ",
		"แสดงงบรายวัน อาหาร",
		"ซ่อนงบรายวัน อาหาร",
		"งบรายวัน",
	}, "\n")
}

func centsToFloat(cents int64) float64 {
	return float64(cents) / 100
}

func emptyFallback(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}
