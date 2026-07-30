package chat

import "time"

// KnowledgeArticle represents a knowledge base document stored with its embedding.
type KnowledgeArticle struct {
	ID             int64                  `json:"id" secure:"encrypt_id"`
	OrganizationID *int64                 `json:"organization_id,omitempty"` // nil = global
	Category       string                 `json:"category"`
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ChunkIndex     int                    `json:"chunk_index"`
	SourceDocID    string                 `json:"source_doc_id"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeleteStatus   int                    `json:"-"`
}

// CreateKnowledgeInput is the input for creating a knowledge article.
type CreateKnowledgeInput struct {
	Category string `json:"category" validate:"required,oneof=module_guide how_to troubleshooting faq policy"`
	Title    string `json:"title" validate:"required,min=5,max=500"`
	Content  string `json:"content" validate:"required,min=20"`
}

// UpdateKnowledgeInput is the input for updating a knowledge article.
type UpdateKnowledgeInput struct {
	Category *string `json:"category" validate:"omitempty,oneof=module_guide how_to troubleshooting faq policy"`
	Title    *string `json:"title" validate:"omitempty,min=5,max=500"`
	Content  *string `json:"content" validate:"omitempty,min=20"`
}

// RetrievedChunk is a similarity search result from the knowledge base.
type RetrievedChunk struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Category   string  `json:"category"`
	Similarity float64 `json:"similarity"`
}
