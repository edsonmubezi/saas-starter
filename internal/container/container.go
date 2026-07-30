package container

import (
	"log"
	"time"

	"github.com/edsonmubezi/myapp/internal/chat"
	"github.com/edsonmubezi/myapp/internal/config"
	"github.com/edsonmubezi/myapp/internal/emailconfig"
	"github.com/edsonmubezi/myapp/internal/importlog"
	"github.com/edsonmubezi/myapp/internal/logs"
	"github.com/edsonmubezi/myapp/internal/organization"
	"github.com/edsonmubezi/myapp/internal/orgsettings"
	"github.com/edsonmubezi/myapp/internal/permission"
	"github.com/edsonmubezi/myapp/internal/platform/alerting"
	"github.com/edsonmubezi/myapp/internal/platform/applog"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/edsonmubezi/myapp/internal/platform/security"
	"github.com/edsonmubezi/myapp/internal/role"
	"github.com/edsonmubezi/myapp/internal/seeder"
	"github.com/edsonmubezi/myapp/internal/user"
	"github.com/edsonmubezi/myapp/pkg/storage"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Container struct {
	UseCases map[string]interface{}
	LogRepo  logs.LogRepository
	Storage  storage.Storage
	DB       *pgxpool.Pool
	UserRepo user.UserRepository

	// Platform Services (PostgreSQL-backed)
	AuditService    *audit.PostgresService
	SecurityService *security.Service
	AlertingService *alerting.Service
	AppLogService   *applog.Service
}

func New(db *pgxpool.Pool, cfg *config.Config) *Container {
	// Initialize storage service (conditionally based on EnableFileUpload)
	var storageService storage.Storage

	if cfg.EnableFileUpload {
		storageCfg := storage.Config{
			Type:        cfg.Storage.Type,
			LocalPath:   cfg.Storage.LocalPath,
			S3Region:    cfg.Storage.S3Region,
			S3Bucket:    cfg.Storage.S3Bucket,
			S3Endpoint:  cfg.Storage.S3Endpoint,
			S3AccessKey: cfg.Storage.S3AccessKey,
			S3SecretKey: cfg.Storage.S3SecretKey,
		}

		var err error
		storageService, err = storage.New(storageCfg)
		if err != nil {
			log.Printf("WARNING: File upload disabled - storage initialization failed: %v", err)
			log.Printf("WARNING: Server will continue with file upload disabled. Set ENABLE_FILE_UPLOAD=false or fix storage configuration.")
			storageService = storage.NewNoOpStorage()
		} else {
			log.Printf("Storage initialized: type=%s", cfg.Storage.Type)
		}
	} else {
		log.Println("File upload feature disabled (ENABLE_FILE_UPLOAD=false)")
		storageService = storage.NewNoOpStorage()
	}

	// ===== REPOSITORIES =====

	// Core repositories
	userRepo := user.NewPostgresUserRepository(db)
	roleRepo := role.NewPostgresRoleRepository(db)
	permissionRepo := permission.NewPostgresPermissionRepository(db)
	organizationRepo := organization.NewPostgresOrganizationRepository(db)
	logRepo := logs.NewPostgresLogRepository(db)

	// Settings repositories
	orgSettingsRepo := orgsettings.NewOrganizationSettingsRepository(db)

	// Import log repository
	importLogRepo := importlog.NewPostgresRepository(db)

	// ===== PLATFORM SERVICES =====
	// Initialize early so they can be injected into use cases

	// Initialize audit service (PostgreSQL-backed)
	auditService := audit.NewPostgresService(db)
	log.Println("Audit service initialized (PostgreSQL)")

	// ===== USE CASES =====

	// Core use cases
	userUseCase := user.NewUserUseCase(userRepo, auditService)
	roleUseCase := role.NewRoleUseCase(roleRepo, auditService)
	permissionUseCase := permission.NewPermissionUseCase(permissionRepo, auditService)
	organizationUseCase := organization.NewOrganizationUseCase(organizationRepo, orgSettingsRepo, auditService)
	// Inject org defaults seeder for auto-seeding new organizations
	orgSeeder := seeder.NewSeedManager(db)
	if orgImpl, ok := organizationUseCase.(*organization.OrganizationImpl); ok {
		orgImpl.SetOrgSeeder(orgSeeder)
	}

	// Import log use case
	importLogUseCase := importlog.NewUseCase(importLogRepo)

	// Organization settings use case
	orgCoreSettingsUseCase := orgsettings.NewOrganizationSettingsUseCase(orgSettingsRepo)

	// Email config use case
	emailConfigRepo := emailconfig.NewRepository(db)
	emailConfigUseCase := emailconfig.NewUseCase(emailConfigRepo)

	// Chat repository, context provider, and use case
	chatRepo := chat.NewPostgresRepository(db)
	knowledgeRepo := chat.NewKnowledgePostgresRepository(db)
	orgContextProvider := chat.NewOrgContextProvider(orgSettingsRepo)
	promptBuilder := chat.NewPromptBuilder(knowledgeRepo, orgContextProvider)
	chatUseCase := chat.NewUseCase(chatRepo, promptBuilder)

	useCases := map[string]interface{}{
		"user":            userUseCase,
		"role":            roleUseCase,
		"permission":      permissionUseCase,
		"organization":    organizationUseCase,
		"logs":            logRepo,
		"orgcoresettings": orgCoreSettingsUseCase,
		"importlog":       importLogUseCase,
		"emailconfig":     emailConfigUseCase,
		"chat":            chatUseCase,
		"knowledgerepo":   knowledgeRepo,
	}

	// ===== REMAINING PLATFORM SERVICES =====

	// Initialize alerting service
	alertingService := alerting.NewService(cfg.Alerting, cfg.Email)
	log.Println("Alerting service initialized")

	// Initialize security event service (PostgreSQL-backed)
	securityService := security.NewService(db, cfg.Alerting.Enabled)
	log.Println("Security event service initialized (PostgreSQL)")

	// Initialize application logging service (PostgreSQL-backed)
	appLogService := applog.NewService(db, applog.Config{
		ServiceName:   "saas-api",
		BufferSize:    1000,
		FlushInterval: 5 * time.Second,
	})
	log.Println("Application logging service initialized (PostgreSQL)")

	return &Container{
		UseCases:        useCases,
		LogRepo:         logRepo,
		Storage:         storageService,
		DB:              db,
		UserRepo:        userRepo,
		AuditService:    auditService,
		SecurityService: securityService,
		AlertingService: alertingService,
		AppLogService:   appLogService,
	}
}
