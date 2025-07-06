package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// loggerKey is the context key for the logger
	loggerKey contextKey = "logger"
)

// Options configures the logger
type Options struct {
	// Level sets the minimum log level
	Level slog.Level
	// Writer is the output destination
	Writer io.Writer
	// AddSource adds source file information to log records
	AddSource bool
	// Format specifies the output format ("text" or "json")
	Format string
}

// DefaultOptions returns default logger options
func DefaultOptions() Options {
	return Options{
		Level:     slog.LevelWarn, // Only show warnings and errors by default
		Writer:    os.Stderr,
		AddSource: false,
		Format:    "text",
	}
}

// New creates a new slog.Logger with the given options
func New(opts Options) *slog.Logger {
	var handler slog.Handler

	handlerOpts := &slog.HandlerOptions{
		Level:     opts.Level,
		AddSource: opts.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time from logs for cleaner CLI output
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}
			// Simplify level output
			if a.Key == slog.LevelKey && len(groups) == 0 {
				level := a.Value.Any().(slog.Level)
				switch {
				case level >= slog.LevelError:
					a.Value = slog.StringValue("ERROR")
				case level >= slog.LevelWarn:
					a.Value = slog.StringValue("WARN")
				case level >= slog.LevelInfo:
					a.Value = slog.StringValue("INFO")
				default:
					a.Value = slog.StringValue("DEBUG")
				}
			}
			return a
		},
	}

	switch opts.Format {
	case "json":
		handler = slog.NewJSONHandler(opts.Writer, handlerOpts)
	default:
		handler = slog.NewTextHandler(opts.Writer, handlerOpts)
	}

	return slog.New(handler)
}

// WithContext returns a new context with the logger attached
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger from the context
// If no logger is found, it returns a default logger
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return New(DefaultOptions())
}

// Info logs at INFO level using logger from context
func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

// Warn logs at WARN level using logger from context
func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

// Error logs at ERROR level using logger from context
func Error(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).ErrorContext(ctx, msg, args...)
}

// Debug logs at DEBUG level using logger from context
func Debug(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).DebugContext(ctx, msg, args...)
}

// With returns a logger with additional attributes
func With(ctx context.Context, args ...any) *slog.Logger {
	return FromContext(ctx).With(args...)
}
