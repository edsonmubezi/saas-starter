package router

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	orgsroutes "github.com/edsonmubezi/myapp/api/router/Orgs"
	superroutes "github.com/edsonmubezi/myapp/api/router/Super"
	"github.com/edsonmubezi/myapp/internal/auth"
	"github.com/edsonmubezi/myapp/internal/config"
	"github.com/edsonmubezi/myapp/internal/emailconfig"
	"github.com/edsonmubezi/myapp/internal/orgsettings"
	"github.com/edsonmubezi/myapp/internal/platform/alerting"
	"github.com/edsonmubezi/myapp/internal/platform/applog"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/edsonmubezi/myapp/internal/platform/email"
	"github.com/edsonmubezi/myapp/internal/platform/security"
	"github.com/edsonmubezi/myapp/internal/user"
	"github.com/edsonmubezi/myapp/pkg/storage"
	"github.com/jackc/pgx/v4/pgxpool"

	_ "github.com/edsonmubezi/myapp/docs"
	"github.com/edsonmubezi/myapp/internal/logs"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRouter(useCases map[string]interface{}, logRepo logs.LogRepository, storageService storage.Storage, db *pgxpool.Pool, auditService *audit.PostgresService, securityService *security.Service, alertingService *alerting.Service, appLogService *applog.Service) *mux.Router {
	r := mux.NewRouter()

	// Initialize storage in handlers
	handler.SetStorage(storageService)

	// Load configuration
	cfg := config.Load()

	// Initialize email resolver (org DB config first, .env fallback)
	initializeEmailResolver(cfg, db)

	// Initialize security repositories for login lockout feature
	initializeSecurityServices(db)

	// Set security service for auth handlers
	handler.SetSecurityService(securityService)

	// ---------- Global Middleware ----------
	r.Use(middleware.TimeoutMiddleware(nil)) // Request timeout with per-route configuration
	r.Use(middleware.SecurityHeadersDefault())
	r.Use(middleware.SanitizeInputs)
	r.Use(middleware.BodyLimit(middleware.DefaultMaxBodySize)) // Prevent OOM via large request bodies

	// ---------- Health Check (for Coolify / Docker) ----------

	// ---------- Swagger ----------
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// ---------- File Uploads (Static) ----------
	// Serve uploaded files with CORS headers for cross-origin access
	uploadsFileServer := http.FileServer(http.Dir("./uploads"))
	serveUploads := func(prefix string) http.Handler {
		return http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			w.Header().Del("X-Frame-Options")
			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			uploadsFileServer.ServeHTTP(w, req)
		}))
	}
	// Primary route: /api/v1/uploads/ — works through nginx API proxy
	r.PathPrefix("/api/v1/uploads/").Handler(serveUploads("/api/v1/uploads/"))
	// Legacy route: /uploads/ — still works for direct backend access
	r.PathPrefix("/uploads/").Handler(serveUploads("/uploads/"))

	// ---------- Public Routes ----------
	// Health check endpoints
	r.HandleFunc("/health", handler.HealthHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/health/detailed", handler.AdvancedHealthHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/health/live", handler.LivenessHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/health/ready", handler.ReadinessHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/health/circuit-breakers", handler.CircuitBreakerStatusHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/health/circuit-breakers/{service}/reset", handler.ResetCircuitBreakerHandler).Methods("POST", "OPTIONS")

	// Authentication routes with rate limiting
	loginRateLimiter := middleware.RateLimitMiddleware(cfg.Security, "login")
	r.Handle("/api/v1/login", loginRateLimiter(http.HandlerFunc(handler.LoginHandler))).Methods("POST", "OPTIONS")
	r.Handle("/api/v1/refresh", loginRateLimiter(http.HandlerFunc(handler.RefreshTokenHandler))).Methods("POST", "OPTIONS")

	// Password reset routes (public) with rate limiting
	r.Handle("/api/v1/auth/forgot-password", loginRateLimiter(http.HandlerFunc(handler.ForgotPasswordHandler))).Methods("POST", "OPTIONS")
	r.Handle("/api/v1/auth/reset-password", loginRateLimiter(http.HandlerFunc(handler.ResetPasswordHandler))).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/auth/verify-reset-token/{token}", handler.VerifyResetTokenHandler).Methods("GET", "OPTIONS")

	// 2FA verification during login (public route with rate limiting)
	r.Handle("/api/v1/auth/2fa/verify-login", loginRateLimiter(http.HandlerFunc(handler.Verify2FALoginHandler))).Methods("POST", "OPTIONS")
	r.Handle("/api/v1/auth/2fa/send-login-code", loginRateLimiter(http.HandlerFunc(handler.Send2FALoginCodeHandler))).Methods("POST", "OPTIONS")

	// ---------- Register Use Cases ----------
	_ = superroutes.RegisterSuperAdminLevelRoutes(nil, useCases)

	// ---------- Second Pass Dependencies ----------
	// Set OrgCoreSettingsUseCase for /users/me endpoint to return organization_type
	if orgCoreSettingsUC, ok := useCases["orgcoresettings"]; ok && orgCoreSettingsUC != nil {
		handler.SetOrgSettingsUseCaseForUserHandler(orgCoreSettingsUC)
	}

	// ---------- Protected Routes ----------
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.SanitizeInputs)
	protected.Use(middleware.JWTMiddleware)
	protected.Use(middleware.LoggingMiddleware(logRepo))
	protected.Use(middleware.RateLimiter(1.0, 2))

	// Authenticated user operations
	protected.HandleFunc("/users/me", handler.GetAuthenticatedUserHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/change-password", handler.ChangePasswordHandler).Methods("PUT")
	protected.HandleFunc("/logout", handler.LogoutHandler).Methods("POST")
	protected.HandleFunc("/auth/verify-password", handler.VerifyPasswordHandler).Methods("POST")
	protected.HandleFunc("/auth/session-config", handler.SessionConfigHandler).Methods("GET")

	// Two-Factor Authentication routes
	protected.HandleFunc("/auth/2fa/status", handler.Get2FAStatusHandler).Methods("GET")
	protected.HandleFunc("/auth/2fa/setup", handler.Setup2FAHandler).Methods("POST")
	protected.HandleFunc("/auth/2fa/verify", handler.Verify2FASetupHandler).Methods("POST")
	protected.HandleFunc("/auth/2fa/disable", handler.Disable2FAHandler).Methods("POST")
	protected.HandleFunc("/auth/2fa/backup-codes/regenerate", handler.RegenerateBackupCodesHandler).Methods("POST")
	protected.HandleFunc("/auth/2fa/send-code", handler.Send2FACodeHandler).Methods("POST")

	// Session management routes
	protected.HandleFunc("/sessions", handler.GetActiveSessionsHandler).Methods("GET")
	protected.HandleFunc("/sessions/count", handler.GetSessionCountHandler).Methods("GET")
	protected.HandleFunc("/sessions/revoke", handler.RevokeSessionHandler).Methods("POST")
	protected.HandleFunc("/sessions/revoke-others", handler.RevokeOtherSessionsHandler).Methods("POST")
	protected.HandleFunc("/sessions/{id}", handler.DeleteSessionHandler).Methods("DELETE")

	// ---------- Isolated Route Groups ----------
	// Each group has its own prefix for complete endpoint isolation
	// This enables future VPN/network-level security per group

	// Admin routes: /api/admin/* (SuperAdmin/HQ only)
	adminRouter := protected.PathPrefix("/admin").Subrouter()
	superroutes.RegisterSuperAdminLevelRoutes(adminRouter, useCases)
	superroutes.RegisterAdminAuditRoutes(adminRouter, auditService)
	superroutes.RegisterAdminAlertingRoutes(adminRouter, alertingService)
	registerSecurityRoutes(adminRouter, securityService) // Security is admin-only
	registerAppLogRoutes(adminRouter, appLogService)     // Logs are admin-only

	// Organization routes: /api/org/* (Tenant/Organization admins)
	orgRouter := protected.PathPrefix("/org").Subrouter()
	orgsroutes.RegisterOrganizationLevelRoutes(orgRouter, useCases)
	orgsroutes.RegisterTenantAuditRoutes(orgRouter, auditService)
	orgsroutes.RegisterTenantAlertingRoutes(orgRouter, alertingService)

	// ---------- Static Files + SPA Fallback ----------
	staticDir := http.Dir("./static")

	// Serve /static/* assets
	r.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(staticDir)))

	// Serve index.html at the root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	}).Methods("GET")

	// SPA fallback: for non-API paths, serve index.html
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/swagger/") ||
			strings.HasPrefix(r.URL.Path, "/uploads/") ||
			strings.HasPrefix(r.URL.Path, "/health") {
			http.NotFound(w, r)
			return
		}

		clean := path.Clean(r.URL.Path)
		candidate := filepath.Join("static", clean)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			http.ServeFile(w, r, candidate)
			return
		}

		http.ServeFile(w, r, "static/index.html")
	})

	return r
}

// initializeEmailResolver creates the email resolver (org DB config first, .env fallback)
// and initializes the reset token service.
func initializeEmailResolver(cfg *config.Config, db *pgxpool.Pool) {
	// Initialize reset token service
	resetTokenService := auth.NewResetTokenService(db)
	handler.SetResetTokenService(resetTokenService)
	log.Println("✓ Reset token service initialized")

	// Build global (fallback) email service from .env
	var globalSvc email.EmailService
	if cfg.Email.Enabled {
		smtpService, err := email.NewSMTPEmailService(email.EmailConfig{
			Enabled:      cfg.Email.Enabled,
			SMTPHost:     cfg.Email.SMTPHost,
			SMTPPort:     cfg.Email.SMTPPort,
			SMTPUser:     cfg.Email.SMTPUser,
			SMTPPassword: cfg.Email.SMTPPassword,
			FromAddress:  cfg.Email.FromAddress,
			FromName:     cfg.Email.FromName,
		})
		if err != nil {
			log.Printf("⚠️  Failed to initialize SMTP email service: %v", err)
			log.Println("   Using mock email service (emails will be logged instead)")
			globalSvc = email.NewMockEmailService()
		} else {
			globalSvc = smtpService
			log.Printf("✓ SMTP fallback email service initialized (host=%s, port=%d)", cfg.Email.SMTPHost, cfg.Email.SMTPPort)
		}
	} else {
		log.Println("  Email service disabled in .env configuration")
		log.Println("   Using mock email service as fallback (emails will be logged instead)")
		globalSvc = email.NewMockEmailService()
	}

	// Create resolver: checks org_email_configs DB table first, falls back to global
	emailCfgRepo := emailconfig.NewRepository(db)
	resolver := email.NewResolver(globalSvc, emailCfgRepo)
	handler.SetEmailResolver(resolver)
	log.Println("✓ Email resolver initialized (org DB config → .env fallback)")
}

// initializeSecurityServices initializes the user repository and login attempts repository for security features
func initializeSecurityServices(db *pgxpool.Pool) {
	// Initialize user repository for security operations (lockout, unlock, etc.)
	userRepo := user.NewPostgresUserRepository(db)
	handler.SetUserRepository(userRepo)
	log.Println("✓ User repository initialized for security features")

	// Initialize login attempts repository for audit trail
	loginAttemptsRepo := auth.NewLoginAttemptsRepository(db)
	handler.SetLoginAttemptsRepository(loginAttemptsRepo)
	log.Println("✓ Login attempts repository initialized")

	// Initialize password reset history repository
	passwordResetHistoryRepo := auth.NewPasswordResetHistoryRepository(db)
	handler.SetPasswordResetHistoryRepo(passwordResetHistoryRepo)
	log.Println("✓ Password reset history repository initialized")

	// Initialize TOTP service for 2FA
	totpService := auth.NewTOTPService(db)
	handler.SetTOTPService(totpService)
	log.Println("✓ TOTP service initialized for two-factor authentication")

	// Initialize org settings repository for session config
	orgSettingsRepo := orgsettings.NewOrganizationSettingsRepository(db)
	handler.SetOrgSettingsRepository(orgSettingsRepo)
	log.Println("✓ Org settings repository initialized for session config")

	// Initialize refresh token repository for session management
	refreshTokenRepo := auth.NewRefreshTokenRepository(db)
	handler.SetRefreshTokenRepository(refreshTokenRepo)
	log.Println("✓ Refresh token repository initialized for session management")
}

// Note: Audit routes have been moved to:
// - Super/audit_routes.go for admin routes (/api/admin/audit/*)
// - Orgs/audit_routes.go for tenant routes (/api/org/audit/*)

// registerSecurityRoutes registers all security event API routes
func registerSecurityRoutes(r *mux.Router, securityService *security.Service) {
	if securityService == nil {
		log.Println("⚠️  Security service not configured, skipping security routes")
		return
	}

	securityHandler := handler.NewSecurityHandler(securityService)

	// Create subrouter with security permission middleware
	// Only HQ admins can view security events
	securityRouter := r.PathPrefix("/security").Subrouter()
	securityRouter.Use(middleware.RequireAnyPermission([]string{"admin.security.view"}))

	// Security dashboard and events
	securityRouter.HandleFunc("/dashboard", securityHandler.GetSecurityDashboard).Methods("GET", "OPTIONS")
	securityRouter.HandleFunc("/events", securityHandler.GetSecurityEvents).Methods("GET", "OPTIONS")
	securityRouter.HandleFunc("/events/{id}", securityHandler.GetSecurityEventByID).Methods("GET", "OPTIONS")

	log.Println("✓ Security routes registered")
}

// Note: Alerting routes have been moved to:
// - Super/alerting_routes.go for admin routes (/api/admin/alerting/*)
// - Orgs/alerting_routes.go for tenant routes (/api/org/alerting/*)

// registerAppLogRoutes registers all application log API routes
func registerAppLogRoutes(r *mux.Router, appLogService *applog.Service) {
	if appLogService == nil {
		log.Println("⚠️  Application log service not configured, skipping log routes")
		return
	}

	appLogHandler := handler.NewAppLogHandler(appLogService)

	// Create subrouter with logs permission middleware
	// Only HQ admins can view system logs
	logsRouter := r.PathPrefix("/logs").Subrouter()
	logsRouter.Use(middleware.RequireAnyPermission([]string{"admin.logs.view"}))

	// Application logs
	logsRouter.HandleFunc("/application", appLogHandler.GetApplicationLogs).Methods("GET", "OPTIONS")
	logsRouter.HandleFunc("/application/{id}", appLogHandler.GetLogByID).Methods("GET", "OPTIONS")

	// Access logs
	logsRouter.HandleFunc("/access", appLogHandler.GetAccessLogs).Methods("GET", "OPTIONS")

	// Log statistics
	logsRouter.HandleFunc("/stats", appLogHandler.GetLogStats).Methods("GET", "OPTIONS")

	// Log metadata
	logsRouter.HandleFunc("/levels", appLogHandler.GetLogLevels).Methods("GET", "OPTIONS")
	logsRouter.HandleFunc("/categories", appLogHandler.GetLogCategories).Methods("GET", "OPTIONS")

	log.Println("✓ Application log routes registered")
}

func getBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return strings.TrimSuffix(baseURL, "/")
}
