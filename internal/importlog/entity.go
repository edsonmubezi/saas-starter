package importlog

// Import status constants
const (
	StatusCompleted = "completed"
	StatusPartial   = "partial"
	StatusFailed    = "failed"
)

// Record type constants
const (
	RecordTypeError     = "error"
	RecordTypeWarning   = "warning"
	RecordTypeDuplicate = "duplicate"
)

// Resolution constants
const (
	ResolutionUpdated = "updated"
	ResolutionSkipped = "skipped"
)

// ImportSession represents a single import attempt.
type ImportSession struct {
	ID            int64  `json:"id" secure:"encrypt_id"`
	ImportType    string `json:"import_type"`
	Status        string `json:"status"`
	TotalRows     int    `json:"total_rows"`
	Successful    int    `json:"successful"`
	Failed        int    `json:"failed"`
	Duplicates    int    `json:"duplicates"`
	ErrorFile     string `json:"error_file,omitempty"`
	CreatedBy     int64  `json:"created_by" secure:"encrypt_id"`
	CreatedByName string `json:"created_by_name"`
	OrgID         int64  `json:"organization_id"`
	CreatedAt     string `json:"created_at"`
}

// ImportFailedRecord represents a single failed or warned record in an import session.
type ImportFailedRecord struct {
	ID              int64  `json:"id" secure:"encrypt_id"`
	ImportSessionID int64  `json:"import_session_id" secure:"encrypt_id"`
	RowNumber       int    `json:"row_number"`
	RecordName      string `json:"record_name"`
	Reason          string `json:"reason"`
	RecordType      string `json:"record_type"`
	ColumnName          string  `json:"column_name,omitempty"`
	Value               string  `json:"value,omitempty"`
	ExistingRecordID    *int64  `json:"existing_record_id,omitempty" secure:"encrypt_id"`
	ResolvedAt          *string `json:"resolved_at,omitempty"`
	Resolution          *string `json:"resolution,omitempty"`
	OrgID               int64   `json:"organization_id"`
}

// DetermineStatus returns the import status based on success/failure counts.
func DetermineStatus(successful, failed int) string {
	if failed == 0 {
		return StatusCompleted
	}
	if successful == 0 {
		return StatusFailed
	}
	return StatusPartial
}
