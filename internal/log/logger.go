// Package logger provides a tiny, dependency-free wrapper around Go's standard
// slog logger. Supports levels, JSON or text output, optional file writing.
package logger

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Level is an alias for slog.Level.
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Options configures the logger.
type Options struct {
	Level      Level
	JSON       bool
	AddSource  bool
	OutputPath string
}

var (
	mu      sync.RWMutex
	base    *slog.Logger
	closers []io.Closer
)

// Init sets up the global logger. Safe to call multiple times.
func Init(opts Options) error {
	mu.Lock()
	defer mu.Unlock()

	for _, c := range closers {
		_ = c.Close()
	}
	closers = nil

	var w io.Writer = os.Stderr
	if p := strings.TrimSpace(opts.OutputPath); p != "" {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		w = f
		closers = append(closers, f)
	}

	var h slog.Handler
	if opts.JSON {
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: opts.Level, AddSource: opts.AddSource})
	} else {
		h = slog.NewTextHandler(w, &slog.HandlerOptions{Level: opts.Level, AddSource: opts.AddSource})
	}

	base = slog.New(h)
	return nil
}

// Default ensures there's always a usable logger.
func Default() *slog.Logger {
	mu.RLock()
	b := base
	mu.RUnlock()
	if b != nil {
		return b
	}
	// Lazy init to stderr, info level, text format.
	_ = Init(Options{Level: LevelInfo, JSON: false, AddSource: true})
	mu.RLock()
	defer mu.RUnlock()
	return base
}

// L returns the current global logger.
func L() *slog.Logger { return Default() }

// With attaches key-value pairs to the global logger and returns a child.
func With(args ...any) *slog.Logger { return Default().With(args...) }

// Named returns a logger with a component name field.
func Named(name string) *slog.Logger { return With("component", name) }

// SetLevel updates the log level at runtime.
// Note: slog’s handlers bake the level at creation, so we rebuild the handler.
func SetLevel(lvl Level) error {
	mu.RLock()
	b := base
	mu.RUnlock()
	if b == nil {
		return errors.New("logger not initialized")
	}
	// For simplicity, just re-init with defaults.
	return Init(Options{Level: lvl, JSON: false, AddSource: true})
}

// Convenience wrappers
func Debug(msg string, args ...any) { Default().Debug(msg, args...) }
func Info(msg string, args ...any)  { Default().Info(msg, args...) }
func Warn(msg string, args ...any)  { Default().Warn(msg, args...) }
func Error(msg string, args ...any) { Default().Error(msg, args...) }

// Sync closes file handles (if any).
func Sync() {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range closers {
		_ = c.Close()
	}
	closers = nil
}
