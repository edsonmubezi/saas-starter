package chat

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// KnowledgeRepository handles knowledge base CRUD and vector search.
type KnowledgeRepository interface {
	CreateArticle(ctx context.Context, article *KnowledgeArticle, embedding []float32) (*KnowledgeArticle, error)
	UpdateArticle(ctx context.Context, id int64, orgID *int64, input *UpdateKnowledgeInput, embedding []float32) (*KnowledgeArticle, error)
	DeleteArticle(ctx context.Context, id int64, orgID *int64) error
	GetArticleByID(ctx context.Context, id int64, orgID *int64) (*KnowledgeArticle, error)
	ListArticles(ctx context.Context, orgID *int64, category string, page, pageSize int) ([]KnowledgeArticle, int, error)
	SearchSimilar(ctx context.Context, queryEmbedding []float32, orgID int64, topK int, minSimilarity float64) ([]RetrievedChunk, error)
	BulkInsert(ctx context.Context, articles []KnowledgeArticle, embeddings [][]float32) error
}

type knowledgePostgresRepo struct {
	db *pgxpool.Pool
}

// NewKnowledgePostgresRepository creates a Postgres-backed KnowledgeRepository.
func NewKnowledgePostgresRepository(db *pgxpool.Pool) KnowledgeRepository {
	return &knowledgePostgresRepo{db: db}
}

func (r *knowledgePostgresRepo) CreateArticle(ctx context.Context, article *KnowledgeArticle, embedding []float32) (*KnowledgeArticle, error) {
	query := `
		INSERT INTO knowledge_embeddings (organization_id, category, title, content, embedding, metadata, chunk_index, source_doc_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		article.OrganizationID,
		article.Category,
		article.Title,
		article.Content,
		embedding,
		article.Metadata,
		article.ChunkIndex,
		article.SourceDocID,
	).Scan(&article.ID, &article.CreatedAt, &article.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create knowledge article: %w", err)
	}
	return article, nil
}

func (r *knowledgePostgresRepo) UpdateArticle(ctx context.Context, id int64, orgID *int64, input *UpdateKnowledgeInput, embedding []float32) (*KnowledgeArticle, error) {
	query := `
		UPDATE knowledge_embeddings
		SET category = COALESCE($1, category),
		    title = COALESCE($2, title),
		    content = COALESCE($3, content),
		    embedding = COALESCE($4, embedding),
		    updated_at = NOW()
		WHERE id = $5 AND delete_status = 0
		  AND (($6::bigint IS NULL AND organization_id IS NULL) OR organization_id = $6)
		RETURNING id, organization_id, category, title, content, chunk_index, source_doc_id, created_at, updated_at`

	var article KnowledgeArticle
	var cat, title, content *string
	if input.Category != nil {
		cat = input.Category
	}
	if input.Title != nil {
		title = input.Title
	}
	if input.Content != nil {
		content = input.Content
	}

	var embeddingParam interface{} = nil
	if embedding != nil {
		embeddingParam = embedding
	}

	err := r.db.QueryRow(ctx, query, cat, title, content, embeddingParam, id, orgID).
		Scan(&article.ID, &article.OrganizationID, &article.Category, &article.Title,
			&article.Content, &article.ChunkIndex, &article.SourceDocID, &article.CreatedAt, &article.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update knowledge article: %w", err)
	}
	return &article, nil
}

func (r *knowledgePostgresRepo) DeleteArticle(ctx context.Context, id int64, orgID *int64) error {
	query := `
		UPDATE knowledge_embeddings SET delete_status = 1, updated_at = NOW()
		WHERE id = $1 AND delete_status = 0
		  AND (($2::bigint IS NULL AND organization_id IS NULL) OR organization_id = $2)`

	_, err := r.db.Exec(ctx, query, id, orgID)
	return err
}

func (r *knowledgePostgresRepo) GetArticleByID(ctx context.Context, id int64, orgID *int64) (*KnowledgeArticle, error) {
	query := `
		SELECT id, organization_id, category, title, content, chunk_index, source_doc_id, created_at, updated_at
		FROM knowledge_embeddings
		WHERE id = $1 AND delete_status = 0
		  AND (($2::bigint IS NULL AND organization_id IS NULL) OR organization_id = $2)`

	var a KnowledgeArticle
	err := r.db.QueryRow(ctx, query, id, orgID).
		Scan(&a.ID, &a.OrganizationID, &a.Category, &a.Title, &a.Content,
			&a.ChunkIndex, &a.SourceDocID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *knowledgePostgresRepo) ListArticles(ctx context.Context, orgID *int64, category string, page, pageSize int) ([]KnowledgeArticle, int, error) {
	countQuery := `
		SELECT COUNT(*) FROM knowledge_embeddings
		WHERE delete_status = 0
		  AND (($1::bigint IS NULL AND organization_id IS NULL) OR organization_id = $1)
		  AND ($2 = '' OR category = $2)`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, orgID, category).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, organization_id, category, title, content, chunk_index, source_doc_id, created_at, updated_at
		FROM knowledge_embeddings
		WHERE delete_status = 0
		  AND (($1::bigint IS NULL AND organization_id IS NULL) OR organization_id = $1)
		  AND ($2 = '' OR category = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, orgID, category, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []KnowledgeArticle
	for rows.Next() {
		var a KnowledgeArticle
		if err := rows.Scan(&a.ID, &a.OrganizationID, &a.Category, &a.Title, &a.Content,
			&a.ChunkIndex, &a.SourceDocID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		articles = append(articles, a)
	}
	return articles, total, nil
}

func (r *knowledgePostgresRepo) SearchSimilar(ctx context.Context, queryEmbedding []float32, orgID int64, topK int, minSimilarity float64) ([]RetrievedChunk, error) {
	query := `
		SELECT id, title, content, category,
		       1 - (embedding <=> $1::vector) AS similarity
		FROM knowledge_embeddings
		WHERE delete_status = 0
		  AND (organization_id IS NULL OR organization_id = $2)
		ORDER BY embedding <=> $1::vector
		LIMIT $3`

	rows, err := r.db.Query(ctx, query, queryEmbedding, orgID, topK)
	if err != nil {
		return nil, fmt.Errorf("knowledge search: %w", err)
	}
	defer rows.Close()

	var chunks []RetrievedChunk
	for rows.Next() {
		var c RetrievedChunk
		if err := rows.Scan(&c.ID, &c.Title, &c.Content, &c.Category, &c.Similarity); err != nil {
			return nil, err
		}
		if c.Similarity >= minSimilarity {
			chunks = append(chunks, c)
		}
	}
	return chunks, nil
}

func (r *knowledgePostgresRepo) BulkInsert(ctx context.Context, articles []KnowledgeArticle, embeddings [][]float32) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, article := range articles {
		_, err := tx.Exec(ctx, `
			INSERT INTO knowledge_embeddings (organization_id, category, title, content, embedding, metadata, chunk_index, source_doc_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			article.OrganizationID,
			article.Category,
			article.Title,
			article.Content,
			embeddings[i],
			article.Metadata,
			article.ChunkIndex,
			article.SourceDocID,
		)
		if err != nil {
			return fmt.Errorf("bulk insert article %d: %w", i, err)
		}
	}

	return tx.Commit(ctx)
}
