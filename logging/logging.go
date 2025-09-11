package logging

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "sync"
    "time"

    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

// LogLevel defines the severity of the log.
type LogLevel string

const (
    LevelInfo  LogLevel = "INFO"
    LevelError LogLevel = "ERROR"
)

// DefaultLogger creates OpenTelemetry span events (in-trace logs).
var DefaultLogger = New()

// JSONLogger writes structured JSON logs to a file for the collector's filelog receiver.
var JSONLogger = NewStructured()

// Logger wraps a span-aware logging API.
type Logger struct{}

// New creates a new Logger.
func New() *Logger { return &Logger{} }

// Info logs a message with INFO level as a span event.
func (l *Logger) Info(ctx context.Context, message string, attrs ...attribute.KeyValue) {
    l.log(ctx, LevelInfo, message, attrs...)
}

// Error logs a message with ERROR level as a span event.
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

// StructuredLogger writes JSON logs to a file.
type StructuredLogger struct {
    mu      sync.Mutex
    f       *os.File
    encoder *json.Encoder
}

// NewStructured creates a JSON logger. The output file defaults to ./app.log
// and can be overridden via APP_LOG_FILE env var.
func NewStructured() *StructuredLogger {
    path := os.Getenv("APP_LOG_FILE")
    if path == "" {
        path = "app.log"
    }
    f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
    if err != nil {
        log.Printf("[WARN] failed to open log file %q: %v", path, err)
        return &StructuredLogger{}
    }
    return &StructuredLogger{f: f, encoder: json.NewEncoder(f)}
}

// Info writes a JSON log with INFO level.
func (l *StructuredLogger) Info(ctx context.Context, message string, attrs ...attribute.KeyValue) {
    l.write(ctx, LevelInfo, message, attrs...)
}

// Error writes a JSON log with ERROR level.
func (l *StructuredLogger) Error(ctx context.Context, message string, attrs ...attribute.KeyValue) {
    l.write(ctx, LevelError, message, attrs...)
}

func (l *StructuredLogger) write(ctx context.Context, level LogLevel, message string, attrs ...attribute.KeyValue) {
    if l.encoder == nil {
        // Fallback if file could not be opened.
        log.Printf("[%s] %s %v", level, message, attrs)
        return
    }
    span := trace.SpanFromContext(ctx)
    sc := span.SpanContext()
    entry := map[string]any{
        "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
        "level":     string(level),
        "message":   message,
        "attributes": attrsToMap(attrs...),
    }
    if sc.IsValid() {
        entry["trace_id"] = sc.TraceID().String()
        entry["span_id"] = sc.SpanID().String()
    }
    l.mu.Lock()
    defer l.mu.Unlock()
    _ = l.encoder.Encode(entry)
}

func attrsToMap(attrs ...attribute.KeyValue) map[string]any {
    m := make(map[string]any, len(attrs))
    for _, a := range attrs {
        switch a.Value.Type() {
        case attribute.STRING:
            m[string(a.Key)] = a.Value.AsString()
        case attribute.INT64:
            m[string(a.Key)] = a.Value.AsInt64()
        case attribute.BOOL:
            m[string(a.Key)] = a.Value.AsBool()
        case attribute.FLOAT64:
            m[string(a.Key)] = a.Value.AsFloat64()
        default:
            m[string(a.Key)] = a.Value.AsString()
        }
    }
    return m
}
