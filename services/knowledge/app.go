// Package knowledge composes PostgreSQL retrieval with an Agent knowledge tool.
package knowledge

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nostosx/agentx-contrib/retrieval"
	"github.com/nostosx/agentx-examples/infrastructure/config"
	"github.com/nostosx/agentx/rag"
	pg "github.com/nostosx/servex-contrib/database/postgres"
	textpg "github.com/nostosx/servex-contrib/search/postgres"
	vectorpg "github.com/nostosx/servex-contrib/vector/pgvector"
)

// DemoEmbedder is a deterministic fixture, not a semantic embedding model.
// Replace it and rebuild document embeddings together for a real application.
type DemoEmbedder struct{}

func (DemoEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = []float32{1, 0, 0}
	}
	return out, nil
}

type Result struct {
	VectorHits int            `json:"vector_hits"`
	TextHits   int            `json:"text_hits"`
	Content    string         `json:"content"`
	Citations  []rag.Citation `json:"citations"`
}

// RunPostgresRAG creates only an explicitly owned demonstration schema. The
// application owns migrations and indexing; retrieval adapters only read. It
// borrows db, leaves demo rows for inspection, and never calls a model API.
func RunPostgresRAG(ctx context.Context, db *sql.DB, cfg config.PostgresRAG) (Result, error) {
	var result Result
	if err := cfg.Validate(); err != nil {
		return result, err
	}
	schema, err := pg.Ident(cfg.Schema)
	if err != nil {
		return result, err
	}
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema); err != nil {
		return result, err
	}
	v, err := vectorpg.New(db, vectorpg.Config{Table: pg.Table{Schema: cfg.Schema, Name: "vectors"}, Namespace: cfg.Namespace, Dimensions: 3})
	if err != nil {
		return result, err
	}
	text, err := textpg.New(db, textpg.Config{Table: pg.Table{Schema: cfg.Schema, Name: "documents"}, Namespace: cfg.Namespace, Engine: textpg.Engine(cfg.SearchEngine)})
	if err != nil {
		return result, err
	}
	if err := v.Migrate(ctx); err != nil {
		return result, err
	}
	if err := text.Migrate(ctx); err != nil {
		return result, err
	}
	chunks := []*rag.Chunk{
		{ID: "postgres", Source: "docs/postgres", Content: "PostgreSQL retrieval combines vector and document search.", Embedding: []float32{1, 0, 0}, Metadata: map[string]any{"kind": "guide"}},
		{ID: "private", Source: "private/notes", Content: "PostgreSQL private document.", Embedding: []float32{1, 0, 0}, Metadata: map[string]any{"kind": "private"}},
	}
	if err := pg.Transaction(ctx, db, func(tx *sql.Tx) error {
		for _, c := range chunks {
			record, err := retrieval.VectorRecord(c)
			if err != nil {
				return err
			}
			doc, err := retrieval.TextDocument(c)
			if err != nil {
				return err
			}
			if err := v.UpsertTx(ctx, tx, record); err != nil {
				return err
			}
			if err := text.UpsertTx(ctx, tx, doc); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return result, err
	}
	vr, err := retrieval.NewServexVector(retrieval.ServexOptions{Vectors: v, Embedder: DemoEmbedder{}})
	if err != nil {
		return result, err
	}
	tr, err := retrieval.NewServexText(retrieval.ServexOptions{Text: text})
	if err != nil {
		return result, err
	}
	filter := rag.Filter{Sources: []string{"docs/postgres"}, Metadata: map[string]any{"kind": "guide"}}
	q := rag.Query{Text: "PostgreSQL", Limit: 3, Filter: filter}
	vectors, err := vr.Retrieve(ctx, q)
	if err != nil {
		return result, err
	}
	documents, err := tr.Retrieve(ctx, q)
	if err != nil {
		return result, err
	}
	var hybridErr error
	hybrid := &rag.Hybrid{Retrievers: []rag.Retriever{vr, tr}, OnError: func(err error) { hybridErr = errors.Join(hybridErr, err) }}
	knowledge := rag.Tool(hybrid, rag.ToolOptions{Filter: filter, Limit: 3, MaxLimit: 3})
	response, err := knowledge.Call(ctx, []byte(`{"query":"PostgreSQL"}`))
	if err != nil {
		return result, err
	}
	if hybridErr != nil {
		return result, hybridErr
	}
	cites, _ := response.Metadata["citations"].([]rag.Citation)
	if response.IsError || len(vectors) != 1 || len(documents) != 1 || len(cites) != 1 || cites[0].Source != "docs/postgres" {
		return result, errors.New("retrieval or citation isolation failed")
	}
	result = Result{VectorHits: len(vectors), TextHits: len(documents), Content: response.Content, Citations: cites}
	return result, nil
}
