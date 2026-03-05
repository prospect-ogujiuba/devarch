CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE embeddings (
    id BIGSERIAL PRIMARY KEY,
    source_type TEXT NOT NULL,
    source_id TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(768) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_type, source_id)
);

CREATE INDEX idx_embeddings_hnsw ON embeddings USING hnsw (embedding vector_cosine_ops);
CREATE INDEX idx_embeddings_source ON embeddings (source_type, source_id);
