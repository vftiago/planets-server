package logger

import (
	"log/slog"
	"os"
	"planets-server/internal/shared/config"
)

func Init() {
	if config.GlobalConfig == nil {
		panic("config must be initialized before logger")
	}

	logConfig := config.GlobalConfig.Logging
	var handler slog.Handler
	
	level := parseLogLevel(logConfig.Level)
	
	if logConfig.JSONFormat {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}
	
	slog.SetDefault(slog.New(handler))
	
	logger := slog.With("component", "logger")
	logger.Debug("Logger initialized",
		"level", logConfig.Level,
		"json_format", logConfig.JSONFormat,
		"environment", config.GlobalConfig.Server.Environment,
	)
}

func parseLogLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
