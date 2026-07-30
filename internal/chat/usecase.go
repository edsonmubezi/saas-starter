package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	secureid "github.com/edsonmubezi/myapp/pkg/encrypt"
	openai "github.com/sashabaranov/go-openai"
)

// UseCase defines the chat business logic interface.
type UseCase interface {
	CreateThreadAndSend(ctx context.Context, orgID, userID int64, input *CreateThreadInput, onChunk StreamCallback) (*Thread, *Message, error)
	SendMessage(ctx context.Context, threadID, orgID, userID int64, input *SendMessageInput, onChunk StreamCallback) (*Message, error)
	ListThreads(ctx context.Context, orgID, userID int64) ([]Thread, error)
	GetMessages(ctx context.Context, threadID, orgID, userID int64) ([]Message, error)
	UpdateThread(ctx context.Context, threadID, orgID, userID int64, input *UpdateThreadInput) error
	DeleteThread(ctx context.Context, threadID, orgID, userID int64) error
	GetAPIKeyConfig(ctx context.Context, orgID int64) (*APIKeyConfig, error)
	GetRawAPIKey(ctx context.Context, orgID int64) (string, error)
	UpdateAPIKey(ctx context.Context, orgID int64, input *APIKeyUpdateInput) error
}

type chatUseCase struct {
	repo          Repository
	promptBuilder *PromptBuilder
}

// NewUseCase creates a new chat use case.
func NewUseCase(repo Repository, promptBuilder *PromptBuilder) UseCase {
	return &chatUseCase{repo: repo, promptBuilder: promptBuilder}
}

func (uc *chatUseCase) CreateThreadAndSend(ctx context.Context, orgID, userID int64, input *CreateThreadInput, onChunk StreamCallback) (*Thread, *Message, error) {
	title := input.Title
	if title == "" {
		// Auto-generate title from first message
		title = input.Message
		if len(title) > 60 {
			title = title[:60] + "..."
		}
	}

	thread, err := uc.repo.CreateThread(ctx, &Thread{
		OrganizationID: orgID,
		UserID:         userID,
		Title:          title,
	})
	if err != nil {
		return nil, nil, err
	}

	msg, err := uc.SendMessage(ctx, thread.ID, orgID, userID, &SendMessageInput{Message: input.Message}, onChunk)
	if err != nil {
		return thread, nil, err
	}
	return thread, msg, nil
}

func (uc *chatUseCase) SendMessage(ctx context.Context, threadID, orgID, userID int64, input *SendMessageInput, onChunk StreamCallback) (*Message, error) {
	// 1. Verify thread ownership
	_, err := uc.repo.GetThread(ctx, threadID, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("thread not found or access denied")
	}

	// 2. Get API key
	config, err := uc.repo.GetAPIKey(ctx, orgID)
	if err != nil || config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured. Please set it in Organization Settings > AI Assistant")
	}

	// 3. Save user message
	userMsg, err := uc.repo.CreateMessage(ctx, &Message{
		ThreadID: threadID,
		Role:     "user",
		Content:  input.Message,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}
	_ = userMsg

	// 4. Load message history (last 20 messages to stay within token limits)
	history, err := uc.repo.GetMessages(ctx, threadID, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}
	if len(history) > 20 {
		history = history[len(history)-20:]
	}

	// 5. Build OpenAI messages with layered system prompt
	systemPrompt := SystemPrompt
	if uc.promptBuilder != nil {
		systemPrompt = uc.promptBuilder.BuildSystemPrompt(ctx, orgID, input.Message, config.APIKey)
	}
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
	}
	for _, msg := range history {
		role := openai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 6. Create streaming request
	model := config.Model
	if model == "" {
		model = openai.GPT4oMini
	}

	client := openai.NewClient(config.APIKey)
	stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   2048,
		Temperature: 0.7,
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start AI stream: %w", err)
	}
	defer stream.Close()

	// 7. Stream chunks
	var fullResponse strings.Builder
	for {
		response, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			// Save partial response if we got any content
			if fullResponse.Len() > 0 {
				uc.repo.CreateMessage(ctx, &Message{
					ThreadID: threadID,
					Role:     "assistant",
					Content:  fullResponse.String(),
				})
			}
			return nil, fmt.Errorf("stream error: %w", recvErr)
		}
		if len(response.Choices) > 0 {
			chunk := response.Choices[0].Delta.Content
			fullResponse.WriteString(chunk)
			if err := onChunk(chunk); err != nil {
				// Client disconnected — save partial response
				if fullResponse.Len() > 0 {
					uc.repo.CreateMessage(ctx, &Message{
						ThreadID: threadID,
						Role:     "assistant",
						Content:  fullResponse.String(),
					})
				}
				return nil, err
			}
		}
	}

	// 8. Save full assistant message
	assistantMsg, err := uc.repo.CreateMessage(ctx, &Message{
		ThreadID: threadID,
		Role:     "assistant",
		Content:  fullResponse.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save assistant response: %w", err)
	}
	return assistantMsg, nil
}

func (uc *chatUseCase) ListThreads(ctx context.Context, orgID, userID int64) ([]Thread, error) {
	return uc.repo.ListThreads(ctx, orgID, userID)
}

func (uc *chatUseCase) GetMessages(ctx context.Context, threadID, orgID, userID int64) ([]Message, error) {
	return uc.repo.GetMessages(ctx, threadID, orgID, userID)
}

func (uc *chatUseCase) UpdateThread(ctx context.Context, threadID, orgID, userID int64, input *UpdateThreadInput) error {
	return uc.repo.UpdateThreadTitle(ctx, threadID, orgID, userID, input.Title)
}

func (uc *chatUseCase) DeleteThread(ctx context.Context, threadID, orgID, userID int64) error {
	return uc.repo.DeleteThread(ctx, threadID, orgID, userID)
}

func (uc *chatUseCase) GetAPIKeyConfig(ctx context.Context, orgID int64) (*APIKeyConfig, error) {
	config, err := uc.repo.GetAPIKey(ctx, orgID)
	if err != nil {
		return &APIKeyConfig{Model: "gpt-4o-mini"}, nil
	}
	// Mask the key for display
	if len(config.APIKey) > 4 {
		config.APIKey = "sk-..." + config.APIKey[len(config.APIKey)-4:]
	}
	return config, nil
}

func (uc *chatUseCase) GetRawAPIKey(ctx context.Context, orgID int64) (string, error) {
	config, err := uc.repo.GetAPIKey(ctx, orgID)
	if err != nil {
		return "", err
	}
	return config.APIKey, nil
}

func (uc *chatUseCase) UpdateAPIKey(ctx context.Context, orgID int64, input *APIKeyUpdateInput) error {
	// Encrypt the API key before storage
	encrypted, err := secureid.EncryptID(input.APIKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt API key: %w", err)
	}
	model := input.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	return uc.repo.UpsertAPIKey(ctx, orgID, encrypted, model)
}
