package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type loggerType int

var (
	text loggerType = 1
	json loggerType = 2
)

func MustNewLogger(name string) *slog.Logger {
	return newLogger(text, name)
}

func MustLoadLoggerJson(name string) *slog.Logger {
	return newLogger(json, name)
}

func newLogger(t loggerType, name string) *slog.Logger {
	logLevel := getLogLevel()

	w, err := setupMultiWriter(logLevel)
	if err != nil {
		panic(fmt.Errorf("failed to open io.Writer: %v", err))
	}

	var (
		handler slog.Handler
		opts    *slog.HandlerOptions = &slog.HandlerOptions{
			AddSource: false,
			Level:     logLevel,
		}
	)

	switch t {
	case text:
		handler = slog.NewTextHandler(w, opts)
	case json:
		handler = slog.NewJSONHandler(w, opts)
	}

	logger := slog.New(handler).With(slog.String("name", name))
	return logger
}

func getLogLevel() slog.Level {
	var logLevel slog.Level = slog.LevelDebug
	logLevelRaw, exists := os.LookupEnv("LOG_LEVEL")
	if !exists {
		return logLevel
	}

	switch strings.ToLower(logLevelRaw) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "error":
		logLevel = slog.LevelError
	case "warn":
		logLevel = slog.LevelWarn
	}

	return logLevel
}

func setupMultiWriter(logLevel slog.Level) (io.Writer, error) {
	if _, err := os.Stat("logs"); err != nil {
		if err := os.Mkdir("logs", 0777); err != nil {
			return nil, err
		}
	}

	logFile, err := os.OpenFile("logs/backend.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	return mw, nil
}
