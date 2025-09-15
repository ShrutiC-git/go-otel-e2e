package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"app/routes"
	"app/tracing"
)

func main() {
	// Initialize OpenTelemetry (traces and metrics).
	shutdown := tracing.InitTracer()

	router := routes.SetupRoutes()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start the server in a goroutine for graceful shutdown.
	go func() {
		log.Println("Server is running on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal and perform graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Perform graceful shutdown of the OTel providers after the server.
	shutdown(ctx)
}
