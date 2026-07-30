package featureflags

import (
	"os"
	"strconv"
	"strings"
)

// Flags contains all feature flags for gradual microservices migration
type Flags struct {
	// Service Migration Flags
	UseIAMService bool // Use standalone IAM service instead of monolith user module
	UseOrgService bool // Use standalone Organization service

	// Feature Flags
	EnableSagaPattern       bool // Use saga pattern for distributed transactions
	EnableEventDriven       bool // Use event-driven communication
	EnableAsyncUserCreation bool // Create users asynchronously via events
}

// Load loads feature flags from environment variables
func Load() Flags {
	return Flags{
		// Service flags (default: false - use monolith)
		UseIAMService: getBoolEnv("FEATURE_IAM_SERVICE", false),
		UseOrgService: getBoolEnv("FEATURE_ORG_SERVICE", false),

		// Feature flags (default: false - keep current behavior)
		EnableSagaPattern:       getBoolEnv("FEATURE_SAGA_PATTERN", false),
		EnableEventDriven:       getBoolEnv("FEATURE_EVENT_DRIVEN", false),
		EnableAsyncUserCreation: getBoolEnv("FEATURE_ASYNC_USER_CREATION", false),
	}
}

// IsAnyServiceEnabled checks if any microservice is enabled
func (f Flags) IsAnyServiceEnabled() bool {
	return f.UseIAMService || f.UseOrgService
}

// getBoolEnv gets a boolean from environment variable with default value
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Handle multiple true values
	value = strings.ToLower(value)
	if value == "true" || value == "1" || value == "yes" || value == "on" {
		return true
	}

	// Handle multiple false values
	if value == "false" || value == "0" || value == "no" || value == "off" {
		return false
	}

	// Parse as boolean
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// Example .env file:
//
// # Microservices Migration Feature Flags
//
// # Phase 1: Foundation (enable saga pattern)
// FEATURE_SAGA_PATTERN=true
// FEATURE_EVENT_DRIVEN=true
//
// # Phase 2: IAM Service (enable after IAM service is deployed)
// FEATURE_IAM_SERVICE=false
// IAM_SERVICE_URL=http://iam-service:8081
//
// # Phase 3: Org Service
// FEATURE_ORG_SERVICE=false
// ORG_SERVICE_URL=http://org-service:8082