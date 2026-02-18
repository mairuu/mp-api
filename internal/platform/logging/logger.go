package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/mairuu/mp-api/internal/platform/observability"
)

func New(logLevel string) *slog.Logger {
	level := slog.LevelInfo
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	base := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	handler := &ContextHandler{Handler: base}
	return slog.New(handler)
}

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if traceID, ok := ctx.Value(observability.TraceIDKey).(string); ok {
			r.AddAttrs(slog.String("trace_id", traceID))
		}
	}

	return h.Handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{
		Handler: h.Handler.WithGroup(name),
	}
}
