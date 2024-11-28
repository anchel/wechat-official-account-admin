package logger

import (
	"log/slog"
	"os"
	"sync/atomic"
)

var root atomic.Value

func init() {
	root.Store(&logger{
		slog.New(DiscardHandler()),
	})
}

func SetDefault(l Logger) {
	// Set default logger
	root.Store(l)
	if lg, ok := l.(*logger); ok {
		slog.SetDefault(lg.inner)
	}
}

func Root() Logger {
	return root.Load().(Logger)
}

func Trace(msg string, ctx ...interface{}) {
	Root().Write(LevelTrace, msg, ctx...)
}

func Debug(msg string, ctx ...interface{}) {
	Root().Write(slog.LevelDebug, msg, ctx...)
}

func Info(msg string, ctx ...interface{}) {
	Root().Write(slog.LevelInfo, msg, ctx...)
}

func Warn(msg string, ctx ...interface{}) {
	Root().Write(slog.LevelWarn, msg, ctx...)
}

func Error(msg string, ctx ...interface{}) {
	Root().Write(slog.LevelError, msg, ctx...)
}

func Crit(msg string, ctx ...interface{}) {
	Root().Write(LevelCrit, msg, ctx...)
	os.Exit(1)
}

func New(ctx ...interface{}) Logger {
	return Root().With(ctx...)
}
