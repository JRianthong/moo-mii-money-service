package bootstrap

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jatuporn/moo-mii-money-service/internal/application/commands"
	"github.com/jatuporn/moo-mii-money-service/internal/application/queries"
	"github.com/jatuporn/moo-mii-money-service/internal/config"
	"github.com/jatuporn/moo-mii-money-service/internal/infrastructure/line"
	"github.com/jatuporn/moo-mii-money-service/internal/infrastructure/persistence/gormrepo"
	httpiface "github.com/jatuporn/moo-mii-money-service/internal/interfaces/http"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewFiberApp(cfg config.Config) (*fiber.App, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	txRepo := gormrepo.NewTransactionRepository(db)
	settingsRepo := gormrepo.NewUserSettingsRepository(db)
	lineClient := line.NewClient(cfg.LineChannelSecret, cfg.LineChannelAccessToken)
	parser := line.NewTextParser(cfg.DefaultCurrency, time.Local)

	app := fiber.New(fiber.Config{
		AppName:      "moo-mii-money-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	httpiface.RegisterRoutes(app, httpiface.Dependencies{
		LineClient:             lineClient,
		TextParser:             parser,
		RecordTransaction:      commands.NewRecordTransactionHandler(txRepo),
		SetBillingCycle:        commands.NewSetBillingCycleHandler(settingsRepo),
		GetBillingCycle:        queries.NewGetBillingCycleHandler(settingsRepo),
		GetMonthlySummary:      queries.NewMonthlySummaryHandler(txRepo, settingsRepo),
		ListRecentTransactions: queries.NewListRecentTransactionsHandler(txRepo),
	})
	return app, nil
}
