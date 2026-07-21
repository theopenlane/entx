// Package logx is a minimal logging helper package standing in for the equivalent
// core-more package (github.com/theopenlane/core/pkg/logx), providing just the surface
// entityops-generated code calls
package logx

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ctxKey struct{}

// WithLogger attaches logger to ctx, retrievable later via FromContext
func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext returns the logger attached to ctx via WithLogger, falling back to the global logger
func FromContext(ctx context.Context) *zerolog.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(zerolog.Logger); ok {
		return &logger
	}

	return &log.Logger
}
