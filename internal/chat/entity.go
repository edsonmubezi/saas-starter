package chat

import "time"

// Thread represents a chat conversation.
type Thread struct {
	ID             int64     `json:"id" secure:"encrypt_id"`
	OrganizationID int64     `json:"organization_id"`
	UserID         int64     `json:"user_id" secure:"encrypt_id"`
	Title          string    `json:"title"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeleteStatus   int       `json:"-"`
	LastMessage    string    `json:"last_message,omitempty"`
	MessageCount   int       `json:"message_count,omitempty"`
}

// Message represents a single chat message.
type Message struct {
	ID        int64     `json:"id" secure:"encrypt_id"`
	ThreadID  int64     `json:"thread_id" secure:"encrypt_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateThreadInput is the input for creating a new thread with an initial message.
type CreateThreadInput struct {
	Title   string `json:"title" validate:"omitempty,max=255"`
	Message string `json:"message" validate:"required,min=1,max=10000"`
}

// SendMessageInput is the input for sending a message to an existing thread.
type SendMessageInput struct {
	Message string `json:"message" validate:"required,min=1,max=10000"`
}

// UpdateThreadInput is the input for updating thread metadata.
type UpdateThreadInput struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}

// APIKeyConfig holds the OpenAI configuration for an org.
type APIKeyConfig struct {
	APIKey string `json:"api_key,omitempty"`
	Model  string `json:"model"`
}

// APIKeyUpdateInput for saving/updating the key.
type APIKeyUpdateInput struct {
	APIKey string `json:"api_key" validate:"required,min=10"`
	Model  string `json:"model" validate:"omitempty,oneof=gpt-4o-mini gpt-4o gpt-4-turbo"`
}

// StreamCallback is called for each text chunk from the LLM.
type StreamCallback func(chunk string) error
