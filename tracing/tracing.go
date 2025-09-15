package tracing

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracer initializes OpenTelemetry and returns a shutdown function.
func InitTracer() func(context.Context) {
	ctx := context.Background()

	// OTel Collector endpoint.
	otlpEndpoint := "localhost:4318"

	// Configure the OTLP HTTP trace exporter (sends traces over HTTP).
	traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(otlpEndpoint), otlptracehttp.WithInsecure())
	if err != nil {
		log.Fatalf("failed to create OTLP trace exporter: %v", err)
	}

	// Configure the OTLP HTTP metric exporter (sends metrics over HTTP).
	metricExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpoint(otlpEndpoint), otlpmetrichttp.WithInsecure())
	if err != nil {
		log.Fatalf("failed to create OTLP metric exporter: %v", err)
	}

	// Define the service resource. These attributes are applied to all telemetry (e.g., for SigNoz).
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("sc-go-app-backend"),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment("development"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// --- Create and set up the Tracer Provider ---
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// --- Create and set up the Meter Provider ---
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// Set the global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Return a shutdown function to be called on application exit.
	return func(ctx context.Context) {
		if err := mp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}
