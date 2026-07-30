package logging

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerKey contextKey = "logger"

// ToContext adds a logger to the context
func ToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves a logger from the context, or returns the global logger
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return GetLogger()
	}

	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}

	return GetLogger()
}

// WithRequestID adds request_id to the logger in context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := FromContext(ctx).With(zap.String("request_id", requestID))
	return ToContext(ctx, logger)
}

// WithTenantID adds tenant_id to the logger in context
func WithTenantID(ctx context.Context, tenantID int64) context.Context {
	logger := FromContext(ctx).With(zap.Int64("tenant_id", tenantID))
	return ToContext(ctx, logger)
}

// WithUserID adds user_id to the logger in context
func WithUserID(ctx context.Context, userID int64) context.Context {
	logger := FromContext(ctx).With(zap.Int64("user_id", userID))
	return ToContext(ctx, logger)
}

// WithOrganizationID adds organization_id to the logger in context
func WithOrganizationID(ctx context.Context, orgID int64) context.Context {
	logger := FromContext(ctx).With(zap.Int64("organization_id", orgID))
	return ToContext(ctx, logger)
}

// WithService adds service name to the logger in context
func WithService(ctx context.Context, service string) context.Context {
	logger := FromContext(ctx).With(zap.String("service", service))
	return ToContext(ctx, logger)
}
