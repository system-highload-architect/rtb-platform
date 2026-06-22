package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New создаёт новый структурированный логгер.
// level: "debug", "info", "warn", "error".
// format: "json" или "text".
// attrs — дополнительные поля, добавляемые к каждой записи (например, slog.String("service", "auction")).
func New(level, format string, attrs ...slog.Attr) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	if len(attrs) > 0 {
		handler = handler.WithAttrs(attrs)
	}

	return slog.New(handler)
}

// parseLevel преобразует строку в slog.Level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
