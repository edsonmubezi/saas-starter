package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer("microfinance-api")
	return tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// RecordError records an error on the span from context
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanAttributes sets multiple attributes on a span
func SetSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// SpanFromContext retrieves the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// Common attribute constructors for convenience
func TenantIDAttr(tenantID int64) attribute.KeyValue {
	return attribute.Int64("tenant_id", tenantID)
}

func UserIDAttr(userID int64) attribute.KeyValue {
	return attribute.Int64("user_id", userID)
}

func ResourceIDAttr(resourceID int64) attribute.KeyValue {
	return attribute.Int64("resource_id", resourceID)
}

func ActionAttr(action string) attribute.KeyValue {
	return attribute.String("action", action)
}

func ResourceTypeAttr(resourceType string) attribute.KeyValue {
	return attribute.String("resource_type", resourceType)
}
