//go:build ignore

package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"{{.ModuleName}}/internal/platform/config"
)

type Bootstrap struct {
	Logger *slog.Logger
	closer io.Closer
}

type contextLoggerKey struct{}

// New 启动时初始化一次 slog；关闭时释放文件句柄。
func New(cfg config.LogConfig, role string) (*Bootstrap, error) {
	writer, err := NewDailyWriter(cfg.Dir, cfg.FilePrefix, role, cfg.RetentionDays)
	if err != nil {
		return nil, err
	}

	output := io.Writer(writer)
	if cfg.Stdout {
		output = io.MultiWriter(os.Stdout, writer)
	}

	options := &slog.HandlerOptions{Level: cfg.Level}
	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, options)
	case "text":
		handler = slog.NewTextHandler(output, options)
	default:
		_ = writer.Close()
		return nil, fmt.Errorf("unsupported log format: %s", cfg.Format)
	}

	return &Bootstrap{
		Logger: slog.New(handler).With("role", role),
		closer: writer,
	}, nil
}

func (b *Bootstrap) Close() error {
	if b == nil || b.closer == nil {
		return nil
	}
	return b.closer.Close()
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if ctx == nil || logger == nil {
		return ctx
	}
	return context.WithValue(ctx, contextLoggerKey{}, logger)
}

func FromContext(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if ctx == nil {
		return fallback
	}
	logger, ok := ctx.Value(contextLoggerKey{}).(*slog.Logger)
	if !ok || logger == nil {
		return fallback
	}
	return logger
}
