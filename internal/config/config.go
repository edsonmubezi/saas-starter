package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment      string // "development", "staging", "production"
	Server           ServerConfig
	Database         DatabaseConfig
	Storage          StorageConfig
	EnableFileUpload bool // Enable/disable file upload feature
	Redis            RedisConfig
	Security         SecurityConfig
	Session          SessionConfig
	Logging          LoggingConfig
	Tracing          TracingConfig
	AuditWorker      AuditWorkerConfig
	AuditArchive     AuditArchiveConfig
	Email            EmailConfig
	Alerting         AlertingConfig
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsStaging returns true if running in staging environment
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// ConnectionString builds a PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

type ServerConfig struct {
	Port string
	Host string
}

type StorageConfig struct {
	Type      string // "local" or "s3"
	LocalPath string

	// S3 Configuration
	S3Region    string
	S3Bucket    string
	S3Endpoint  string // Optional: for custom endpoints (MinIO)
	S3AccessKey string
	S3SecretKey string

	// PostgreSQL Backup Configuration
	PGBackupEnabled bool
	PGBackupBucket  string
	PGBackupPrefix  string
}

type LoggingConfig struct {
	Level    string
	Format   string
	Sampling bool
}

type TracingConfig struct {
	Enabled      bool
	OTLPEndpoint string
	ServiceName  string
}

type AuditWorkerConfig struct {
	Enabled      bool
	PollInterval time.Duration
	BatchSize    int
}

type AuditArchiveConfig struct {
	Enabled      bool
	S3Bucket     string
	CronSchedule string
}

type EmailConfig struct {
	Enabled      bool
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

type AlertingConfig struct {
	Enabled         bool
	SlackWebhookURL string
	TeamsWebhookURL string
	WebhookURL      string
	WebhookSecret   string
	// SMS provider settings
	SMSProvider string
	SMSAPIKey   string
	SMSUsername string
	SMSSenderID string
}

type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Password string
	DB       int
}

type SecurityConfig struct {
	// Password Policy
	PasswordMinLength      int
	PasswordRequireUpper   bool
	PasswordRequireLower   bool
	PasswordRequireNumber  bool
	PasswordRequireSpecial bool
	PasswordForbidCommon   bool

	// JWT Configuration
	JWTSecretMinLength int

	// Rate Limiting
	RateLimitEnabled       bool
	RateLimitLoginAttempts int
	RateLimitLoginWindow   time.Duration

	// SuperAdmin Configuration
	SuperAdminCrossOrgAccess bool
}

type SessionConfig struct {
	Enabled             bool
	MaxConcurrent       int
	TokenExpiry         time.Duration
	RefreshTokenExpiry  time.Duration
	RotateRefreshTokens bool
}

func Load() *Config {
	return &Config{
		Environment: getEnv("APP_ENV", "development"), // default to development
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getIntEnv("DB_PORT", 5432),
			User:     getEnv("DB_USER", "microfinance"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "microfinance"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		EnableFileUpload: getBoolEnv("ENABLE_FILE_UPLOAD", false),
		Storage: StorageConfig{
			Type:      autoDetectStorageType(),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),

			S3Region:    getEnv("AWS_REGION", ""),
			S3Bucket:    getEnv("AWS_S3_BUCKET", ""),
			S3Endpoint:  getEnv("AWS_S3_ENDPOINT", ""),
			S3AccessKey: getEnv("AWS_ACCESS_KEY_ID", ""),
			S3SecretKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),

			PGBackupEnabled: getBoolEnv("PGBACKUP_ENABLED", false),
			PGBackupBucket:  getEnv("PGBACKUP_S3_BUCKET", ""),
			PGBackupPrefix:  getEnv("PGBACKUP_S3_PREFIX", "wal-archives/"),
		},
		Logging: LoggingConfig{
			Level:    getEnv("LOG_LEVEL", "info"),
			Format:   getEnv("LOG_FORMAT", "json"),
			Sampling: getBoolEnv("LOG_SAMPLING", true),
		},
		Tracing: TracingConfig{
			Enabled:      getBoolEnv("OTEL_ENABLED", true),
			OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo:4317"),
			ServiceName:  getEnv("OTEL_SERVICE_NAME", "microfinance-api"),
		},
		AuditWorker: AuditWorkerConfig{
			Enabled:      getBoolEnv("AUDIT_WORKER_ENABLED", true),
			PollInterval: time.Duration(getIntEnv("AUDIT_WORKER_POLL_INTERVAL", 5)) * time.Second,
			BatchSize:    getIntEnv("AUDIT_WORKER_BATCH_SIZE", 500),
		},
		AuditArchive: AuditArchiveConfig{
			Enabled:      getBoolEnv("AUDIT_ARCHIVE_ENABLED", false),
			S3Bucket:     getEnv("AUDIT_S3_BUCKET", ""),
			CronSchedule: getEnv("AUDIT_ARCHIVE_CRON", "0 0 * * *"), // Daily at midnight
		},
		Email: EmailConfig{
			Enabled:      getBoolEnv("SMTP_ENABLED", false),
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getIntEnv("SMTP_PORT", 587),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromAddress:  getEnv("SMTP_FROM_ADDRESS", "noreply@microfinance.com"),
			FromName:     getEnv("SMTP_FROM_NAME", "Microfinance"),
		},
		Redis: RedisConfig{
			Enabled:  getBoolEnv("REDIS_ENABLED", true),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getIntEnv("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Security: SecurityConfig{
			PasswordMinLength:      getIntEnv("PASSWORD_MIN_LENGTH", 12),
			PasswordRequireUpper:   getBoolEnv("PASSWORD_REQUIRE_UPPER", true),
			PasswordRequireLower:   getBoolEnv("PASSWORD_REQUIRE_LOWER", true),
			PasswordRequireNumber:  getBoolEnv("PASSWORD_REQUIRE_NUMBER", true),
			PasswordRequireSpecial: getBoolEnv("PASSWORD_REQUIRE_SPECIAL", true),
			PasswordForbidCommon:   getBoolEnv("PASSWORD_FORBID_COMMON", true),

			JWTSecretMinLength: getIntEnv("JWT_SECRET_MIN_LENGTH", 64),

			RateLimitEnabled:       getBoolEnv("RATE_LIMIT_ENABLED", true),
			RateLimitLoginAttempts: getIntEnv("RATE_LIMIT_LOGIN_ATTEMPTS", 20),
			RateLimitLoginWindow:   time.Duration(getIntEnv("RATE_LIMIT_LOGIN_WINDOW_MINUTES", 15)) * time.Minute,

			SuperAdminCrossOrgAccess: getBoolEnv("SUPERADMIN_CROSS_ORG_ACCESS", false),
		},
		Session: SessionConfig{
			Enabled:             getBoolEnv("SESSION_MANAGEMENT_ENABLED", true),
			MaxConcurrent:       getIntEnv("SESSION_MAX_CONCURRENT", 5),
			TokenExpiry:         time.Duration(getIntEnv("SESSION_TOKEN_EXPIRY_MINUTES", 30)) * time.Minute, // 30 minutes for security
			RefreshTokenExpiry:  time.Duration(getIntEnv("SESSION_REFRESH_TOKEN_EXPIRY_DAYS", 7)) * 24 * time.Hour,
			RotateRefreshTokens: getBoolEnv("SESSION_ROTATE_REFRESH_TOKENS", true),
		},
		Alerting: AlertingConfig{
			Enabled:         getBoolEnv("ALERTING_ENABLED", true),
			SlackWebhookURL: getEnv("SLACK_WEBHOOK_URL", ""),
			TeamsWebhookURL: getEnv("TEAMS_WEBHOOK_URL", ""),
			WebhookURL:      getEnv("ALERTING_WEBHOOK_URL", ""),
			WebhookSecret:   getEnv("ALERTING_WEBHOOK_SECRET", ""),
			SMSProvider:     getEnv("SMS_PROVIDER", ""),
			SMSAPIKey:       getEnv("SMS_API_KEY", ""),
			SMSUsername:     getEnv("SMS_USERNAME", ""),
			SMSSenderID:     getEnv("SMS_SENDER_ID", "Microfinance"),
		},
	}
}

func (c *Config) Address() string {
	return ":" + c.Server.Port
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// autoDetectStorageType intelligently determines storage type:
// 1. If STORAGE_TYPE is explicitly set, use it
// 2. If AWS S3 credentials are configured, use S3
// 3. Otherwise, default to local storage
func autoDetectStorageType() string {
	// First, check if explicitly set
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		return storageType
	}

	// Auto-detect: if AWS credentials are present, assume S3
	awsBucket := os.Getenv("AWS_S3_BUCKET")
	awsRegion := os.Getenv("AWS_REGION")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")

	if awsBucket != "" && awsRegion != "" && awsAccessKey != "" {
		return "s3"
	}

	// Default to local storage for development
	return "local"
}