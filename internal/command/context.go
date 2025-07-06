package command

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

// Context provides shared resources to commands
type Context struct {
	Writer io.Writer
	Logger *slog.Logger
	ctx    context.Context
}

// WithContext returns the underlying context.Context
func (c *Context) Context() context.Context {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	return c.ctx
}

// SetContext sets the underlying context.Context
func (c *Context) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// Printf writes formatted output to the user
func (c *Context) Printf(format string, args ...interface{}) {
	fmt.Fprintf(c.Writer, format, args...)
}

// Println writes a line to the user
func (c *Context) Println(args ...interface{}) {
	fmt.Fprintln(c.Writer, args...)
}

// LogInfo logs an informational message
func (c *Context) LogInfo(msg string, args ...any) {
	if c.Logger != nil {
		c.Logger.InfoContext(c.Context(), msg, args...)
	}
}

// LogWarn logs a warning message
func (c *Context) LogWarn(msg string, args ...any) {
	if c.Logger != nil {
		c.Logger.WarnContext(c.Context(), msg, args...)
	}
}

// LogError logs an error message
func (c *Context) LogError(msg string, args ...any) {
	if c.Logger != nil {
		c.Logger.ErrorContext(c.Context(), msg, args...)
	}
}

// LogDebug logs a debug message
func (c *Context) LogDebug(msg string, args ...any) {
	if c.Logger != nil {
		c.Logger.DebugContext(c.Context(), msg, args...)
	}
}
