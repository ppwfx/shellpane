package logutil

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type contextLoggerKey struct{}

// MustLoggerValue returns the logger from the context and panics if the logger is not in the context
func MustLoggerValue(ctx context.Context) (logger *zap.SugaredLogger) {
	logger, ok := ctx.Value(contextLoggerKey{}).(*zap.SugaredLogger)
	if !ok {
		panic("expected logger")
	}

	return
}

// LoggerValue returns the logger from the context
func LoggerValue(ctx context.Context) (logger *zap.SugaredLogger, err error) {
	logger, ok := ctx.Value(contextLoggerKey{}).(*zap.SugaredLogger)
	if !ok {
		return logger, errors.Errorf("failed to get logger: expected *zap.SugaredLogger, got=%v", logger)
	}

	return
}

// WithLoggerValue adds the logger to the context
func WithLoggerValue(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, contextLoggerKey{}, logger)
}
