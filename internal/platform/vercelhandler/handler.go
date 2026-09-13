package vercelhandler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/jatuporn/moo-mii-money-service/internal/bootstrap"
	"github.com/jatuporn/moo-mii-money-service/internal/config"
)

var (
	appOnce sync.Once
	app     *fiber.App
	appErr  error
)

func Serve(w http.ResponseWriter, r *http.Request) {
	appOnce.Do(func() {
		cfg, err := config.Load()
		if err != nil {
			appErr = err
			return
		}
		app, appErr = bootstrap.NewFiberApp(cfg)
	})
	if appErr != nil {
		log.Printf("build app: %v", appErr)
		http.Error(w, "service is not configured", http.StatusInternalServerError)
		return
	}
	adaptor.FiberApp(app)(w, r)
}
