// api/middleware/cors.go
package middleware

import (
	"os"
	"strings"

	"net/http"

	"github.com/rs/cors"
)

// CORSHandler wraps the next handler with CORS configuration based on APP_ENV.
func CORSHandler(next http.Handler) http.Handler {
	// Load env vars (set in Coolify or .env)
	appEnv := strings.ToLower(os.Getenv("RUNTIME_ENV"))
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		if appEnv == "production" {
			origin = "https://app.overseerhr.online" // Default prod origin
		} else {
			origin = "http://localhost:5174" // Default dev origin
		}
	}

	// Allow multiple origins if needed (comma-separated in env, e.g. for staging)
	allowedOrigins := strings.Split(origin, ",")
	for i, o := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(o)
	}

	// Debug mode: Enable in dev for logging
	// debug := appEnv != "production"

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,                                                                    // Dynamic: dev=localhost, prod=your domain
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},                      // Your existing
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Requested-With", "Accept", "Origin"}, // Your existing
		ExposedHeaders:   []string{"Content-Length", "Content-Disposition"},                                 // Expose Content-Disposition for file downloads
		AllowCredentials: true,                                                                              // Your existing
		Debug:            false,                                                                             // Logs in dev, silent in prod
		MaxAge:           86400,                                                                             // Optional: Cache preflight for 24h (add if needed)
	})

	return c.Handler(next)
}
