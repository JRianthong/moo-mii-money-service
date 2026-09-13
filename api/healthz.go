package handler

import (
	"net/http"

	"github.com/jatuporn/moo-mii-money-service/internal/platform/vercelhandler"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	vercelhandler.Serve(w, r)
}
