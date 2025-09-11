package handlers

import (
    "encoding/json"
    "math/rand/v2"
    "net/http"
    "time"

    "go.opentelemetry.io/otel/attribute"

    "app/logging"
)

// InventoryResponse is the JSON response payload for the inventory check.
type InventoryResponse struct {
    Status  string `json:"status"`
    Message string `json:"message"`
    DelayMS int    `json:"delay_ms"`
}

// CheckInventoryHandler responds with a success message and a simulated delay.
func CheckInventoryHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    delay := rand.IntN(601) + 200

    // Simulate downstream latency (e.g., a database call).
    time.Sleep(time.Duration(delay) * time.Millisecond)

    resp := InventoryResponse{
        Status:  "success",
        Message: "Inventory checked successfully",
        DelayMS: delay,
    }

    // Add structured logs with the simulated delay.
    logging.DefaultLogger.Info(ctx, "Inventory checked successfully", attribute.Int("inventory.check.delay_ms", delay))
    logging.JSONLogger.Info(ctx, "Inventory checked successfully", attribute.Int("inventory.check.delay_ms", delay))

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        logging.DefaultLogger.Error(ctx, "Error encoding inventory response", attribute.String("error.reason", err.Error()))
        logging.JSONLogger.Error(ctx, "Error encoding inventory response", attribute.String("error.reason", err.Error()))
    }

}
