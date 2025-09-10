package handlers

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"app/logging"
)

type InventoryResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	DelayMS int    `json:"delay_ms"`
}

func CheckInventoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	delay := rand.IntN(601) + 200

	// Pause execution for the calculated delay duration for a downstream (perhaps a DB call)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	resp := InventoryResponse{
		Status:  "success",
		Message: "Inventory checked successfully",
		DelayMS: delay,
	}

	// Add structured log with the simulated delay.
	logging.DefaultLogger.Info(ctx, "Inventory checked successfully", attribute.Int("inventory.check.delay_ms", delay))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.DefaultLogger.Error(ctx, "Error encoding inventory response", attribute.String("error.reason", err.Error()))
	}

}
