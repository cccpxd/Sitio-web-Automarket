package logging

import (
	"log/slog"
	"os"
	"time"

	"automarket/internal/config"
)

// New builds a structured logger that writes to stdout, which is the only
// durable log sink available in Vercel Functions.
func New(cfg config.Config) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				if value, ok := attr.Value.Any().(time.Time); ok {
					attr.Value = slog.TimeValue(value.UTC())
				}
			}
			return attr
		},
	})
	return slog.New(handler).With(
		"service", "automarket",
		"environment", cfg.Environment,
	)
}
