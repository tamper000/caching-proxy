package logger

import (
	"log/slog"
	"os"

	"github.com/tamper000/caching-proxy/internal/models"
)

var (
	file *os.File
)

func NewLogger(config models.Logger) {
	var logLevel slog.Level

	switch config.Level {
	case "INFO":
		logLevel = slog.LevelInfo
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "ERROR":
		logLevel = slog.LevelError
	}

	logFile, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("File could not be opened", "error", err)
		os.Exit(1)
	}

	file = logFile

	handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)
}

func CloseFile() {
	file.Close()
}
