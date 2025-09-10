package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"app/logging"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type OrderResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

const (
	instrumentationName = "app/handlers"
	statusSuccess       = "success"
	statusFailure       = "failure"
)

var (
	// Get a meter from the global meter provider
	meter = otel.Meter(instrumentationName)
	// Create the counter instrument
	ordersProcessedCounter metric.Int64Counter
)

func init() {
	var err error
	ordersProcessedCounter, err = meter.Int64Counter(
		"orders_processed_total",
		metric.WithDescription("The total number of orders processed"),
		metric.WithUnit("{order}"),
	)
	if err != nil {
		// If a critical component like a metric counter fails, the application should not start.
		log.Fatalf("failed to create orders_processed_total counter: %v", err)
	}
}

// CreateOrderHandler with a 10% failure-rate
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the current context and a tracer.
	// This ctx contains the parent span from the otelhttp middleware.
	ctx := r.Context()
	tracer := otel.Tracer(instrumentationName)

	// Simulate some initial processing latency (e.g., request validation, business logic).
	time.Sleep(time.Duration(rand.IntN(50)+30) * time.Millisecond)

	// Simulate a 10% chance of failure.
	if rand.IntN(10) == 0 {
		// From the 10% of failures, make half of them database errors.
		if rand.IntN(2) == 0 {
			handleDBError(w, r, tracer)
		} else {
			handlePaymentError(w, r, tracer)
		}
		return // Stop processing after handling the error.
	}

	// --- Success Path ---

	// Database Operation
	dbCtx, dbSpan := tracer.Start(ctx, "db_process_order")
	defer dbSpan.End()
	time.Sleep(time.Duration(rand.IntN(100)+50) * time.Millisecond) // Simulate DB work
	dbSpan.SetStatus(codes.Ok, "order processed successfully")

	// Increment the counter with a "success" status attribute.
	ordersProcessedCounter.Add(dbCtx, 1, metric.WithAttributes(attribute.String("status", statusSuccess)))

	orderID := rand.IntN(1000)
	resp := OrderResponse{
		Status:  "success",
		Message: "Order created successfully",
	}

	trace.SpanFromContext(ctx).SetStatus(codes.Ok, "order created successfully")
	logging.DefaultLogger.Info(ctx, "Order created successfully", attribute.Int("order.id", orderID))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.DefaultLogger.Error(ctx, "Error encoding response", attribute.String("error.reason", err.Error()))
	}
}

// handleDBError simulates a database-related failure. It creates a span for the DB
// operation, marks it as an error, and then returns a 500 Internal Server Error.
func handleDBError(w http.ResponseWriter, r *http.Request, tracer trace.Tracer) {
	ctx := r.Context()
	dbCtx, dbSpan := tracer.Start(ctx, "db_process_order")
	defer dbSpan.End()

	// Simulate a short delay for the failed DB attempt.
	time.Sleep(time.Duration(rand.IntN(40)+10) * time.Millisecond)

	err := errors.New("simulated database constraint violation")
	handleRequestError(dbCtx, dbSpan, "database operation failed", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// handlePaymentError simulates a payment processing failure. It creates a span for the
// payment operation, marks it as an error, and returns a 500 Internal Server Error.
func handlePaymentError(w http.ResponseWriter, r *http.Request, tracer trace.Tracer) {
	ctx := r.Context()
	paymentCtx, paymentSpan := tracer.Start(ctx, "process_payment")
	defer paymentSpan.End()

	err := errors.New("simulated payment provider error")
	handleRequestError(paymentCtx, paymentSpan, "payment processing failed", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// handleRequestError is a helper function to centralize the logic for instrumenting
// an error. It logs the error, increments the failure metric, and sets the span status.
func handleRequestError(ctx context.Context, span trace.Span, message string, err error) {
	logging.DefaultLogger.Error(ctx, message, attribute.String("error.reason", err.Error()))
	ordersProcessedCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("status", statusFailure)))
	span.RecordError(err)
	span.SetStatus(codes.Error, message)
	// Mark the parent span (from the otelhttp middleware) as failed.
	trace.SpanFromContext(ctx).SetStatus(codes.Error, message)
}
