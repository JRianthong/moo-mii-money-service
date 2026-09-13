package http

import (
	"context"
	"fmt"
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
	GetMonthlySummary      queries.MonthlySummaryHandler
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
		return deps.LineClient.ReplyText(ctx, event.ReplyToken, "อ่านยอดไม่สำเร็จ ลองพิมพ์เช่น: จ่าย 120 ข้าวกลางวัน")
	}

	var reply string
	switch intent.Type {
	case line.ParsedIntentRecord:
		result, err := deps.RecordTransaction.Handle(ctx, intent.Command)
		if err != nil {
			return err
		}
		reply = formatRecorded(result.Transaction, result.AlreadyRecorded)
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

func formatRecorded(tx money.Transaction, alreadyRecorded bool) string {
	label := "รายจ่าย"
	if tx.Type() == money.TransactionTypeIncome {
		label = "รายรับ"
	}
	prefix := "บันทึก"
	if alreadyRecorded {
		prefix = "เคยบันทึกแล้ว"
	}
	return fmt.Sprintf("%s%s %.2f %s\n%s", prefix, label, tx.Amount().Float64(), tx.Amount().Currency(), emptyFallback(tx.Note().String(), tx.Category().String()))
}

func formatSummary(summary queries.MonthlySummaryResult) string {
	return fmt.Sprintf(
		"สรุปเดือน %s\nรายรับ %.2f\nรายจ่าย %.2f\nคงเหลือ %.2f",
		summary.From.Format("01/2006"),
		centsToFloat(summary.IncomeCents),
		centsToFloat(summary.ExpenseCents),
		centsToFloat(summary.BalanceCents),
	)
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
