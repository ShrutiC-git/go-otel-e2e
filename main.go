package main

import (
	"fmt"
	"log"
	"net/http"

	"app/routes"
	"app/tracing"
)

func main() {
	// Initializing OpenTel Tracer
	shutdown := tracing.InitTracer()
	defer shutdown()

	router := routes.SetupRoutes()

	fmt.Println("Server is running on port 8080")
	// The router now has instrumentation applied at the route level.
	log.Fatal(http.ListenAndServe(":8080", router))
}
