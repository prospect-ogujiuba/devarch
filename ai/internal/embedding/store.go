package embedding

import (
	"database/sql"
	"fmt"
	"strings"
)

const Dimensions = 768

type Document struct {
	ID         int64   `json:"id"`
	SourceType string  `json:"source_type"`
	SourceID   string  `json:"source_id"`
	Content    string  `json:"content"`
	Score      float64 `json:"score,omitempty"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Upsert(sourceType, sourceID, content string, vec []float32) error {
	vecStr := pgVector(vec)
	_, err := s.db.Exec(`
		INSERT INTO embeddings (source_type, source_id, content, embedding)
		VALUES ($1, $2, $3, $4::vector)
		ON CONFLICT (source_type, source_id)
		DO UPDATE SET content = EXCLUDED.content, embedding = EXCLUDED.embedding, updated_at = NOW()
	`, sourceType, sourceID, content, vecStr)
	return err
}

func (s *Store) Search(vec []float32, limit int) ([]Document, error) {
	vecStr := pgVector(vec)
	rows, err := s.db.Query(`
		SELECT id, source_type, source_id, content, 1 - (embedding <=> $1::vector) AS score
		FROM embeddings
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vecStr, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.SourceType, &d.SourceID, &d.Content, &d.Score); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *Store) DeleteBySource(sourceType, sourceID string) error {
	_, err := s.db.Exec(`DELETE FROM embeddings WHERE source_type = $1 AND source_id = $2`, sourceType, sourceID)
	return err
}

func pgVector(vec []float32) string {
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%g", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
