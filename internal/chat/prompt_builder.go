package chat

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// PromptBuilder assembles the multi-layer system prompt for OpenAI calls.
type PromptBuilder struct {
	knowledgeRepo   KnowledgeRepository
	contextProvider *OrgContextProvider
}

// NewPromptBuilder creates a new PromptBuilder.
// knowledgeRepo can be nil if RAG is not yet enabled.
func NewPromptBuilder(knowledgeRepo KnowledgeRepository, contextProvider *OrgContextProvider) *PromptBuilder {
	return &PromptBuilder{
		knowledgeRepo:   knowledgeRepo,
		contextProvider: contextProvider,
	}
}

// BuildSystemPrompt assembles the complete system prompt from all layers:
//   - Layer 1: Enhanced static prompt (always present)
//   - Layer 2: Dynamic org context (departments, positions, etc. — cached)
//   - Layer 3: RAG chunks (relevant knowledge articles from vector search)
func (b *PromptBuilder) BuildSystemPrompt(ctx context.Context, orgID int64, userMessage string, apiKey string) string {
	var parts []string

	// Layer 1: Static prompt
	parts = append(parts, SystemPrompt)

	// Layer 2: Dynamic org context
	if b.contextProvider != nil {
		if orgCtx := b.contextProvider.GetOrgContext(ctx, orgID); orgCtx != "" {
			parts = append(parts, orgCtx)
		}
	}

	// Layer 3: RAG retrieval (requires knowledgeRepo + apiKey)
	if b.knowledgeRepo != nil && apiKey != "" {
		if ragCtx := b.retrieveRAGContext(ctx, orgID, userMessage, apiKey); ragCtx != "" {
			parts = append(parts, ragCtx)
		}
	}

	return strings.Join(parts, "\n")
}

// retrieveRAGContext embeds the user query and searches for similar knowledge chunks.
func (b *PromptBuilder) retrieveRAGContext(ctx context.Context, orgID int64, query string, apiKey string) string {
	// Create embedding service with the org's API key
	embeddingSvc := NewEmbeddingService(apiKey)

	// Embed the user query
	queryEmbedding, err := embeddingSvc.Embed(ctx, query)
	if err != nil {
		log.Printf("chat: RAG embedding failed: %v", err)
		return "" // graceful degradation
	}

	// Search for similar chunks (top 3, minimum similarity 0.72)
	chunks, err := b.knowledgeRepo.SearchSimilar(ctx, queryEmbedding, orgID, 3, 0.72)
	if err != nil {
		log.Printf("chat: RAG search failed: %v", err)
		return ""
	}
	if len(chunks) == 0 {
		return ""
	}

	// Format retrieved chunks
	var sections []string
	for _, chunk := range chunks {
		sections = append(sections, fmt.Sprintf("### %s\n%s", chunk.Title, chunk.Content))
	}

	return "\n\n## Relevant Knowledge Base Articles\nUse the following reference material to help answer the user's question:\n\n" +
		strings.Join(sections, "\n\n")
}
