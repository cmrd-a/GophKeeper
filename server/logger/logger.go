package logger

import (
	"log/slog"
	"os"
	"strings"
)

func GetLogLevelFromEnv(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func NewLogger() (*slog.Logger, *slog.LevelVar) {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)

	slg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
	return slg, lvl
}
