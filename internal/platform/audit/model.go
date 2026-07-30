package audit

import (
	"time"
)

// Severity levels for audit events
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// ActorType defines who performed the action
type ActorType string

const (
	ActorUser    ActorType = "user"
	ActorService ActorType = "service"
	ActorSystem  ActorType = "system"
)

// AuditEvent represents a single audit trail entry
type AuditEvent struct {
	// Unique identifier
	AuditID string `json:"audit_id"`

	// When it happened
	OccurredAt time.Time `json:"occurred_at"`

	// Who did it
	ActorType ActorType `json:"actor_type"`
	ActorID   *int64    `json:"actor_id,omitempty"`
	TenantID  int64     `json:"tenant_id"`

	// What was done
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   *int64 `json:"resource_id,omitempty"`

	// Request context
	RequestID     string `json:"request_id"`
	IP            string `json:"ip"`
	UserAgent     string `json:"user_agent"`
	OriginService string `json:"origin_service"`

	// Severity
	Severity Severity `json:"severity"`

	// State changes (optional)
	BeforeJSON map[string]interface{} `json:"before_json,omitempty"`
	AfterJSON  map[string]interface{} `json:"after_json,omitempty"`

	// Versioning
	SchemaVersion int    `json:"schema_version"`
	SigSHA256     []byte `json:"sig_sha256,omitempty"`
}

// AuditDocument represents an audit event for PostgreSQL storage and queries
type AuditDocument struct {
	AuditID      string                 `json:"audit_id"`
	Timestamp    time.Time              `json:"timestamp"`
	TenantID     int64                  `json:"tenant_id"`
	ActorType    ActorType              `json:"actor_type"`
	ActorID      *int64                 `json:"actor_id,omitempty"`
	ActorEmail   string                 `json:"actor_email,omitempty"`
	ActorName    string                 `json:"actor_name,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *int64                 `json:"resource_id,omitempty"`
	ResourceName string                 `json:"resource_name,omitempty"`
	Severity     Severity               `json:"severity"`
	RequestID    string                 `json:"request_id,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Endpoint     string                 `json:"endpoint,omitempty"`
	BeforeState  map[string]interface{} `json:"before_state,omitempty"`
	AfterState   map[string]interface{} `json:"after_state,omitempty"`
	Changes      []FieldChange          `json:"changes,omitempty"`
	Signature    string                 `json:"signature,omitempty"`
}

// FieldChange represents a single field that changed
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old"`
	NewValue interface{} `json:"new"`
}

// AuditFilters for querying audit events
type AuditFilters struct {
	TenantID     int64      `json:"tenant_id"`
	ResourceType *string    `json:"resource_type,omitempty"`
	ResourceID   *int64     `json:"resource_id,omitempty"`
	ActorID      *int64     `json:"actor_id,omitempty"`
	ActorSearch  *string    `json:"actor_search,omitempty"`
	Action       *string    `json:"action,omitempty"`
	Severity     *Severity  `json:"severity,omitempty"`
	FromDate     *time.Time `json:"from_date,omitempty"`
	ToDate       *time.Time `json:"to_date,omitempty"`
}

// OutboxEntry represents an event in the outbox table
type OutboxEntry struct {
	ID          int64      `json:"id"`
	EventJSON   []byte     `json:"event_json"`
	CreatedAt   time.Time  `json:"created_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	RetryCount  int        `json:"retry_count"`
	LastError   *string    `json:"last_error,omitempty"`
}

// Standard actions for common operations
const (
	// Authentication
	ActionLoginSuccess  = "LOGIN_SUCCESS"
	ActionLoginFailed   = "LOGIN_FAILED"
	ActionLogout        = "LOGOUT"
	ActionPasswordReset = "PASSWORD_RESET"
	ActionMFAEnabled    = "MFA_ENABLED"
	ActionMFADisabled   = "MFA_DISABLED"

	// CRUD operations
	ActionCreated = "CREATED"
	ActionUpdated = "UPDATED"
	ActionDeleted = "DELETED"
	ActionViewed  = "VIEWED"

	// Authorization
	ActionRoleGranted  = "ROLE_GRANTED"
	ActionRoleRevoked  = "ROLE_REVOKED"
	ActionPermGranted  = "PERMISSION_GRANTED"
	ActionPermRevoked  = "PERMISSION_REVOKED"

	// Data operations
	ActionExported = "EXPORTED"
	ActionImported = "IMPORTED"
	ActionArchived = "ARCHIVED"
	ActionRestored = "RESTORED"

	// Employees
	ActionEmployeeCreated    = "EMPLOYEE_CREATED"
	ActionEmployeeUpdated    = "EMPLOYEE_UPDATED"
	ActionEmployeeDeleted    = "EMPLOYEE_DELETED"
	ActionEmployeePromoted   = "EMPLOYEE_PROMOTED"
	ActionEmployeeTerminated = "EMPLOYEE_TERMINATED"

	// Contract Status Transitions
	ActionEmployeeApproved   = "EMPLOYEE_APPROVED"
	ActionProbationExtended  = "PROBATION_EXTENDED"
	ActionProbationFinished  = "PROBATION_FINISHED"
	ActionContractExtended   = "CONTRACT_EXTENDED"
	ActionEmployeeSuspended  = "EMPLOYEE_SUSPENDED"
	ActionEmployeeReinstated = "EMPLOYEE_REINSTATED"
	ActionEmployeeRetained   = "EMPLOYEE_RETAINED"
	ActionContractAutoExpired = "CONTRACT_AUTO_EXPIRED"

	// Loans
	ActionLoanCreated   = "LOAN_CREATED"
	ActionLoanApproved  = "LOAN_APPROVED"
	ActionLoanRejected  = "LOAN_REJECTED"
	ActionLoanDisbursed = "LOAN_DISBURSED"
	ActionLoanRepaid    = "LOAN_REPAID"
	ActionLoanDefaulted = "LOAN_DEFAULTED"
	ActionLoanCompleted = "LOAN_COMPLETED"

	// Generic workflow actions
	ActionApproved  = "APPROVED"
	ActionRejected  = "REJECTED"
	ActionDisbursed = "DISBURSED"
	ActionCompleted = "COMPLETED"

	// Payroll
	ActionPayrollGenerated = "PAYROLL_GENERATED"
	ActionPayrollApproved  = "PAYROLL_APPROVED"
	ActionPayrollExported  = "PAYROLL_EXPORTED"
	ActionPaymentProcessed = "PAYMENT_PROCESSED"
)

// Resource types
const (
	ResourceEmployee      = "Employee"
	ResourceUser          = "User"
	ResourceRole          = "Role"
	ResourcePermission    = "Permission"
	ResourceLoan          = "Loan"
	ResourcePayroll       = "Payroll"
	ResourceOrganization  = "Organization"
	ResourceDepartment    = "Department"
	ResourceDesignation   = "Designation"
	ResourcePosition      = "Position"
	ResourceLeaveRequest  = "LeaveRequest"
	ResourceAttendance    = "Attendance"
	ResourceDocument      = "Document"
	ResourcePayrollItem   = "PayrollItem"

	// Performance Management
	ResourceRatingScale       = "RatingScale"
	ResourceCompetency        = "Competency"
	ResourceKPI               = "KPI"
	ResourceGoal              = "Goal"
	ResourceReviewCycle       = "ReviewCycle"
	ResourceReviewTemplate    = "ReviewTemplate"
	ResourcePerformanceReview = "PerformanceReview"
	ResourceFeedback360       = "Feedback360"
	ResourcePIP               = "PIP"
)

// Performance Management actions
const (
	// Review Cycle actions
	ActionCycleActivated = "CYCLE_ACTIVATED"
	ActionCycleCompleted = "CYCLE_COMPLETED"
	ActionReviewsInitiated = "REVIEWS_INITIATED"

	// Performance Review actions
	ActionSelfReviewSubmitted    = "SELF_REVIEW_SUBMITTED"
	ActionManagerReviewSubmitted = "MANAGER_REVIEW_SUBMITTED"
	ActionReviewCalibrated       = "REVIEW_CALIBRATED"
	ActionReviewCompleted        = "REVIEW_COMPLETED"

	// 360 Feedback actions
	ActionFeedbackRequested = "FEEDBACK_REQUESTED"
	ActionFeedbackSubmitted = "FEEDBACK_SUBMITTED"

	// PIP actions
	ActionPIPCreated   = "PIP_CREATED"
	ActionPIPUpdated   = "PIP_UPDATED"
	ActionPIPCheckIn   = "PIP_CHECKIN"
	ActionPIPCompleted = "PIP_COMPLETED"

	// Bonus actions
	ActionBonusSynced = "BONUS_SYNCED"
)
