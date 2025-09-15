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
	// Meter from the global meter provider.
	meter = otel.Meter(instrumentationName)
	// Counter for processed orders.
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
		// Fatal: required metric instrument could not be created.
		log.Fatalf("failed to create orders_processed_total counter: %v", err)
	}
}

// CreateOrderHandler simulates a 10% failure rate.
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {

	// Get the current context and a tracer.
	// The context contains the parent span from the otelhttp middleware.
	ctx := r.Context()
	tracer := otel.Tracer(instrumentationName)

	// Simulate initial processing latency (e.g., validation, business logic).
	time.Sleep(time.Duration(rand.IntN(50)+30) * time.Millisecond)

	// Decide if this request should fail (10% chance).
	if rand.IntN(10) == 0 {
		// Half of failures occur during the database step.
		if rand.IntN(2) == 0 {
			handleDBError(w, r, tracer)
			return
		}

		// Otherwise, the DB step succeeds but payment fails next.
		_, dbSpan := tracer.Start(ctx, "db.insert_order")
		time.Sleep(time.Duration(rand.IntN(100)+50) * time.Millisecond)
		dbSpan.SetStatus(codes.Ok, "order record inserted")
		dbSpan.End()

		// Now fail during payment.
		handlePaymentError(w, r, tracer)
		return
	}

	// --- Success Path ---

	// Database step
	_, dbSpan := tracer.Start(ctx, "db.insert_order")
	time.Sleep(time.Duration(rand.IntN(100)+50) * time.Millisecond) // Simulate DB work
	dbSpan.SetStatus(codes.Ok, "order record inserted")
	dbSpan.End()

	// Payment step
	_, paySpan := tracer.Start(ctx, "payment.process")
	time.Sleep(time.Duration(rand.IntN(80)+40) * time.Millisecond) // Simulate payment work
	paySpan.SetStatus(codes.Ok, "payment processed successfully")
	paySpan.End()

	// Increment the counter with a "success" status attribute after the workflow.
	ordersProcessedCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("status", statusSuccess)))

	// Prepare and send the response.
	orderID := rand.IntN(1000)
	resp := OrderResponse{
		Status:  "success",
		Message: "Order created successfully",
	}

	// Parent trace POST /createOder
	trace.SpanFromContext(ctx).SetStatus(codes.Ok, "order created successfully")
	logging.DefaultLogger.Info(ctx, "Order created successfully", attribute.Int("order.id", orderID))
	logging.JSONLogger.Info(ctx, "Order created successfully", attribute.Int("order.id", orderID))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.DefaultLogger.Error(ctx, "Error encoding response", attribute.String("error.reason", err.Error()))
		logging.JSONLogger.Error(ctx, "Error encoding response", attribute.String("error.reason", err.Error()))
	}
}

// handleDBError simulates a database-related failure. It creates a span for the
// DB operation, marks it as an error, and returns HTTP 500.
func handleDBError(w http.ResponseWriter, r *http.Request, tracer trace.Tracer) {
	ctx := r.Context()
	dbCtx, dbSpan := tracer.Start(ctx, "db.insert_order")
	// Simulate a short delay for the failed DB attempt.
	time.Sleep(time.Duration(rand.IntN(40)+10) * time.Millisecond)

	err := errors.New("simulated database constraint violation")
	handleRequestError(dbCtx, dbSpan, "database operation failed", err, "database")
	dbSpan.End()
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// handlePaymentError simulates a payment processing failure. It creates a span for
// the payment operation, marks it as an error, and returns HTTP 500.
func handlePaymentError(w http.ResponseWriter, r *http.Request, tracer trace.Tracer) {
	ctx := r.Context()
	paymentCtx, paymentSpan := tracer.Start(ctx, "payment.process")
	err := errors.New("simulated payment provider error")
	handleRequestError(paymentCtx, paymentSpan, "payment processing failed", err, "payment")
	paymentSpan.End()
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// handleRequestError centralizes error instrumentation: logs, metric, and span status.
func handleRequestError(ctx context.Context, span trace.Span, message string, err error, stage string) {
	logging.DefaultLogger.Error(ctx, message,
		attribute.String("error.stage", stage),
		attribute.String("error.reason", err.Error()),
	)
	logging.JSONLogger.Error(ctx, message,
		attribute.String("error.stage", stage),
		attribute.String("error.reason", err.Error()),
	)
	ordersProcessedCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("status", statusFailure)))
	span.RecordError(err)
	span.SetStatus(codes.Error, message)
	// Mark the request span (from otelhttp) as failed.
	trace.SpanFromContext(ctx).SetStatus(codes.Error, message)
}
