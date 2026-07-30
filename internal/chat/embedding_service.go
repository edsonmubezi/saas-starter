package chat

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

// EmbeddingService generates vector embeddings using OpenAI.
type EmbeddingService interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

type openAIEmbeddingService struct {
	client *openai.Client
}

// NewEmbeddingService creates an embedding service using the given OpenAI API key.
func NewEmbeddingService(apiKey string) EmbeddingService {
	return &openAIEmbeddingService{
		client: openai.NewClient(apiKey),
	}
}

func (s *openAIEmbeddingService) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}

func (s *openAIEmbeddingService) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: texts,
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, err
	}
	results := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		results[i] = d.Embedding
	}
	return results, nil
}
