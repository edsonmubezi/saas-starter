package chat

import (
	"context"
	"fmt"
	"time"

	secureid "github.com/edsonmubezi/myapp/pkg/encrypt"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository defines the data access interface for chat.
type Repository interface {
	CreateThread(ctx context.Context, thread *Thread) (*Thread, error)
	ListThreads(ctx context.Context, orgID, userID int64) ([]Thread, error)
	GetThread(ctx context.Context, id, orgID, userID int64) (*Thread, error)
	UpdateThreadTitle(ctx context.Context, id, orgID, userID int64, title string) error
	DeleteThread(ctx context.Context, id, orgID, userID int64) error
	CreateMessage(ctx context.Context, msg *Message) (*Message, error)
	GetMessages(ctx context.Context, threadID, orgID, userID int64) ([]Message, error)
	GetAPIKey(ctx context.Context, orgID int64) (*APIKeyConfig, error)
	UpsertAPIKey(ctx context.Context, orgID int64, encryptedKey, model string) error
}

type postgresRepo struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new chat repository.
func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) CreateThread(ctx context.Context, thread *Thread) (*Thread, error) {
	query := `INSERT INTO chat_threads (organization_id, user_id, title)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, user_id, title, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, thread.OrganizationID, thread.UserID, thread.Title)
	t := &Thread{}
	err := row.Scan(&t.ID, &t.OrganizationID, &t.UserID, &t.Title, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create thread: %w", err)
	}
	return t, nil
}

func (r *postgresRepo) ListThreads(ctx context.Context, orgID, userID int64) ([]Thread, error) {
	query := `SELECT t.id, t.title, t.created_at, t.updated_at,
		COALESCE((SELECT content FROM chat_messages WHERE thread_id = t.id ORDER BY created_at DESC LIMIT 1), '') AS last_message,
		(SELECT COUNT(*) FROM chat_messages WHERE thread_id = t.id) AS message_count
		FROM chat_threads t
		WHERE t.organization_id = $1 AND t.user_id = $2 AND t.delete_status = 0
		ORDER BY t.updated_at DESC`
	rows, err := r.db.Query(ctx, query, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("list threads: %w", err)
	}
	defer rows.Close()

	var threads []Thread
	for rows.Next() {
		var t Thread
		if err := rows.Scan(&t.ID, &t.Title, &t.CreatedAt, &t.UpdatedAt, &t.LastMessage, &t.MessageCount); err != nil {
			return nil, fmt.Errorf("scan thread: %w", err)
		}
		t.OrganizationID = orgID
		t.UserID = userID
		// Truncate last message preview
		if len(t.LastMessage) > 100 {
			t.LastMessage = t.LastMessage[:100] + "..."
		}
		threads = append(threads, t)
	}
	return threads, nil
}

func (r *postgresRepo) GetThread(ctx context.Context, id, orgID, userID int64) (*Thread, error) {
	query := `SELECT id, title, created_at, updated_at
		FROM chat_threads
		WHERE id = $1 AND organization_id = $2 AND user_id = $3 AND delete_status = 0`
	t := &Thread{OrganizationID: orgID, UserID: userID}
	err := r.db.QueryRow(ctx, query, id, orgID, userID).Scan(&t.ID, &t.Title, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get thread: %w", err)
	}
	return t, nil
}

func (r *postgresRepo) UpdateThreadTitle(ctx context.Context, id, orgID, userID int64, title string) error {
	query := `UPDATE chat_threads SET title = $1, updated_at = $2
		WHERE id = $3 AND organization_id = $4 AND user_id = $5 AND delete_status = 0`
	tag, err := r.db.Exec(ctx, query, title, time.Now(), id, orgID, userID)
	if err != nil {
		return fmt.Errorf("update thread: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("thread not found")
	}
	return nil
}

func (r *postgresRepo) DeleteThread(ctx context.Context, id, orgID, userID int64) error {
	query := `UPDATE chat_threads SET delete_status = 1, updated_at = $1
		WHERE id = $2 AND organization_id = $3 AND user_id = $4 AND delete_status = 0`
	tag, err := r.db.Exec(ctx, query, time.Now(), id, orgID, userID)
	if err != nil {
		return fmt.Errorf("delete thread: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("thread not found")
	}
	return nil
}

func (r *postgresRepo) CreateMessage(ctx context.Context, msg *Message) (*Message, error) {
	query := `INSERT INTO chat_messages (thread_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING id, thread_id, role, content, created_at`
	m := &Message{}
	err := r.db.QueryRow(ctx, query, msg.ThreadID, msg.Role, msg.Content).
		Scan(&m.ID, &m.ThreadID, &m.Role, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	// Update thread's updated_at
	r.db.Exec(ctx, `UPDATE chat_threads SET updated_at = NOW() WHERE id = $1`, msg.ThreadID)
	return m, nil
}

func (r *postgresRepo) GetMessages(ctx context.Context, threadID, orgID, userID int64) ([]Message, error) {
	query := `SELECT m.id, m.thread_id, m.role, m.content, m.created_at
		FROM chat_messages m
		JOIN chat_threads t ON t.id = m.thread_id
		WHERE m.thread_id = $1 AND t.organization_id = $2 AND t.user_id = $3 AND t.delete_status = 0
		ORDER BY m.created_at ASC`
	rows, err := r.db.Query(ctx, query, threadID, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ThreadID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *postgresRepo) GetAPIKey(ctx context.Context, orgID int64) (*APIKeyConfig, error) {
	query := `SELECT COALESCE(openai_api_key, ''), COALESCE(openai_model, 'gpt-4o-mini')
		FROM organization_settings
		WHERE organization_id = $1`
	config := &APIKeyConfig{}
	err := r.db.QueryRow(ctx, query, orgID).Scan(&config.APIKey, &config.Model)
	if err != nil {
		return nil, fmt.Errorf("get api key: %w", err)
	}
	// Decrypt the key if present
	if config.APIKey != "" {
		decrypted, decErr := secureid.DecryptID(config.APIKey)
		if decErr == nil {
			config.APIKey = decrypted
		}
	}
	return config, nil
}

func (r *postgresRepo) UpsertAPIKey(ctx context.Context, orgID int64, encryptedKey, model string) error {
	query := `UPDATE organization_settings SET openai_api_key = $1, openai_model = $2
		WHERE organization_id = $3`
	tag, err := r.db.Exec(ctx, query, encryptedKey, model, orgID)
	if err != nil {
		return fmt.Errorf("upsert api key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("organization settings not found")
	}
	return nil
}
