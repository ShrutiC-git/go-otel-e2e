package logging

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// LogLevel defines the severity of the log.
type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelError LogLevel = "ERROR"
)

// DefaultLogger is a pre-configured global logger that creates OpenTelemetry span events.
var DefaultLogger = New()

// Logger wraps a span-aware logging API.
type Logger struct{}

// New creates a new Logger.
func New() *Logger {
	return &Logger{}
}

// Info logs a message with INFO level.
func (l *Logger) Info(ctx context.Context, message string, attrs ...attribute.KeyValue) {
	l.log(ctx, LevelInfo, message, attrs...)
}

// Error logs a message with ERROR level.
func (l *Logger) Error(ctx context.Context, message string, attrs ...attribute.KeyValue) {
	l.log(ctx, LevelError, message, attrs...)
}

// log records the message as a span event if a span exists in the context.
// If no span is found, it falls back to the standard Go logger.
func (l *Logger) log(ctx context.Context, level LogLevel, message string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		// No span in context, fallback to standard logger.
		log.Printf("[%s] %s %v", level, message, attrs)
		return
	}

	// Pre-allocate slice for efficiency.
	allAttrs := make([]attribute.KeyValue, 0, len(attrs)+2)
	allAttrs = append(allAttrs, attribute.String("log.level", string(level)))
	allAttrs = append(allAttrs, attribute.String("log.message", message))
	allAttrs = append(allAttrs, attrs...)

	// Record the log as a span event.
	span.AddEvent("log", trace.WithAttributes(allAttrs...))
}
