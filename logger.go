package rfc9457

import (
	"log/slog"
)

var logger *slog.Logger

func Logger() *slog.Logger {
	EnsureLogger()
	return logger
}

func SetLogger(l *slog.Logger) {
	logger = l
}

func EnsureLogger() *slog.Logger {
	if logger == nil {
		panic("Must call rfc9457.SetLogger() with a *slog.Logger before reaching this check.")
	}
	return logger
}
