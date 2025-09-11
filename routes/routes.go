package routes

import (
	"net/http"

	"app/handlers"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// SetupRoutes defines all the application's routes and maps them to their corresponding handlers.
func SetupRoutes() *http.ServeMux {
	router := http.NewServeMux()

	// Wrap each handler with otelhttp.NewHandler to create a distinct span for each route.
	// The second argument to NewHandler sets the span name.
    createOrderHandler := otelhttp.NewHandler(http.HandlerFunc(handlers.CreateOrderHandler), "POST /createOrder")
	router.Handle("/createOrder", createOrderHandler)

	checkInventoryHandler := otelhttp.NewHandler(http.HandlerFunc(handlers.CheckInventoryHandler), "GET /checkInventory")
	router.Handle("/checkInventory", checkInventoryHandler)

	return router
}
