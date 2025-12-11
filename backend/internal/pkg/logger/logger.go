package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional context methods
type Logger struct {
	*zap.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string
	Format     string // json or console
	Output     string // stdout or file path
	TimeFormat string
}

// contextKey is a custom type for context keys
type contextKey string

const (
	// LoggerKey is the key for storing logger in context
	LoggerKey contextKey = "logger"
	// RequestIDKey is the key for request ID
	RequestIDKey contextKey = "request_id"
	// TraceIDKey is the key for trace ID
	TraceIDKey contextKey = "trace_id"
)

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	level := parseLevel(cfg.Level)

	// Quyết định hình hài của log (JSON hay Text?)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		// TEXT Format
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Ghi log vào đâu? (File hay Console?)
	var writer zapcore.WriteSyncer
	if cfg.Output == "stdout" || cfg.Output == "" {
		writer = zapcore.AddSync(os.Stdout)
	} else {
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writer = zapcore.AddSync(file)
	}

	core := zapcore.NewCore(encoder, writer, level)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{Logger: logger}, nil
}

// parseLevel parses string level to zapcore.Level
// Ghi log cấp độ nào? (Debug, Info...?)
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// WithContext creates a new logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := []zap.Field{}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if len(fields) == 0 {
		return l
	}

	return &Logger{Logger: l.With(fields...)}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{Logger: l.With(zap.Any(key, value))}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]any) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{Logger: l.With(zapFields...)}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.With(zap.Error(err))}
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return logger
	}
	// Return a no-op logger if not found
	return &Logger{Logger: zap.NewNop()}
}

// ToContext adds logger to context
func ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Helper functions to avoid importing zap in other packages

// String alias for zap.String
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int alias for zap.Int
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Duration alias for zap.Duration
func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

// Any alias for zap.Any
func Any(key string, val any) zap.Field {
	return zap.Any(key, val)
}
