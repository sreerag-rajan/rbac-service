package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var (
	defaultLogger *slog.Logger
	serviceName   = "RBAC SERVICE"
)

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{} // Remove default time, we might want custom format or just let it be
			}
			return a
		},
	}
	// Using JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
}

func SetServiceName(name string) {
	serviceName = name
}

func logMsg(ctx context.Context, level Level, msg string, data interface{}, tags []string) {
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	case LevelFatal:
		slogLevel = slog.LevelError // slog doesn't have Fatal, we'll exit manually
	}

	// Get caller info
	pc, file, line, ok := runtime.Caller(2)
	funcName := "unknown"
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName = fn.Name()
		}
	}

	attrs := []slog.Attr{
		slog.String("service", serviceName),
		slog.String("file", file),
		slog.String("func", funcName),
		slog.Int("line", line),
		slog.Any("data", data),
		slog.Any("tags", tags),
		slog.Time("timestamp", time.Now()),
	}

	defaultLogger.LogAttrs(ctx, slogLevel, msg, attrs...)

	if level == LevelFatal {
		os.Exit(1)
	}
}

func Debug(ctx context.Context, msg string, data interface{}, tags ...string) {
	logMsg(ctx, LevelDebug, msg, data, tags)
}

func Info(ctx context.Context, msg string, data interface{}, tags ...string) {
	logMsg(ctx, LevelInfo, msg, data, tags)
}

func Warn(ctx context.Context, msg string, data interface{}, tags ...string) {
	logMsg(ctx, LevelWarn, msg, data, tags)
}

func Error(ctx context.Context, msg string, data interface{}, tags ...string) {
	logMsg(ctx, LevelError, msg, data, tags)
}

func Fatal(ctx context.Context, msg string, data interface{}, tags ...string) {
	logMsg(ctx, LevelFatal, msg, data, tags)
}
