CREATE TABLE knowledge_embeddings (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NULL REFERENCES organizations(id),
    category        VARCHAR(100) NOT NULL DEFAULT 'general',
    title           VARCHAR(500) NOT NULL,
    content         TEXT NOT NULL,
    embedding       vector(1536) NOT NULL,
    metadata        JSONB DEFAULT '{}',
    chunk_index     INT NOT NULL DEFAULT 0,
    source_doc_id   VARCHAR(255) DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delete_status   INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_knowledge_embeddings_vector
    ON knowledge_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE INDEX idx_knowledge_embeddings_org
    ON knowledge_embeddings(organization_id, delete_status);

CREATE INDEX idx_knowledge_embeddings_category
    ON knowledge_embeddings(category, delete_status);
