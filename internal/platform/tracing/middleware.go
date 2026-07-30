package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPMiddleware wraps handlers with OpenTelemetry instrumentation
func HTTPMiddleware(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "http-server",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
}

// HTTPClient returns an HTTP client with tracing instrumentation
func HTTPClient() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
