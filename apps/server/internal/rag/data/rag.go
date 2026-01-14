package data

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	internaldata "github.com/ZTH7/RAGDesk/apps/server/internal/data"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/rag/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type ragRepo struct {
	log        *log.Helper
	db         *sql.DB
	vector     *qdrantSearchClient
	collection string
}

// NewRAGRepo creates a new rag repo.
func NewRAGRepo(data *internaldata.Data, cfg *conf.Data, logger log.Logger) biz.RAGRepo {
	collection := ""
	var vector *qdrantSearchClient
	if cfg != nil && cfg.Vectordb != nil && cfg.Vectordb.Driver == "qdrant" && cfg.Vectordb.Endpoint != "" {
		vector = newQdrantSearchClient(cfg.Vectordb.Endpoint, cfg.Vectordb.ApiKey)
		collection = cfg.Vectordb.Collection
	}
	if strings.TrimSpace(collection) == "" {
		collection = "ragdesk_chunks"
	}
	return &ragRepo{
		log:        log.NewHelper(logger),
		db:         data.DB,
		vector:     vector,
		collection: collection,
	}
}

func (r *ragRepo) ResolveBotKnowledgeBases(ctx context.Context, botID string) ([]biz.BotKnowledgeBase, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	botID = strings.TrimSpace(botID)
	if botID == "" {
		return nil, nil
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT kb_id, priority, weight
		FROM bot_kb WHERE tenant_id = ? AND bot_id = ? ORDER BY priority DESC, created_at DESC`,
		tenantID,
		botID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.BotKnowledgeBase, 0)
	for rows.Next() {
		var item biz.BotKnowledgeBase
		if err := rows.Scan(&item.KBID, &item.Priority, &item.Weight); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ragRepo) Search(ctx context.Context, req biz.VectorSearchRequest) ([]biz.VectorSearchResult, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	if r.vector == nil {
		return nil, kerrors.InternalServer("VECTORDB_MISSING", "vectordb not configured")
	}
	kbID := strings.TrimSpace(req.KBID)
	if kbID == "" {
		return nil, nil
	}
	points, err := r.vector.Search(ctx, r.collection, req.Vector, req.TopK, tenantID, kbID, req.ScoreThreshold)
	if err != nil {
		return nil, err
	}
	out := make([]biz.VectorSearchResult, 0, len(points))
	for _, p := range points {
		out = append(out, biz.VectorSearchResult{
			ChunkID:           p.ChunkID,
			DocumentID:        p.DocumentID,
			DocumentVersionID: p.DocumentVersionID,
			KBID:              p.KBID,
			Score:             p.Score,
		})
	}
	return out, nil
}

func (r *ragRepo) LoadChunks(ctx context.Context, chunkIDs []string) (map[string]biz.ChunkMeta, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	if len(chunkIDs) == 0 {
		return map[string]biz.ChunkMeta{}, nil
	}
	query, args := buildChunkQuery(tenantID, chunkIDs)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]biz.ChunkMeta, len(chunkIDs))
	for rows.Next() {
		var meta biz.ChunkMeta
		if err := rows.Scan(
			&meta.ChunkID,
			&meta.KBID,
			&meta.DocumentID,
			&meta.DocumentVersionID,
			&meta.Content,
			&meta.Section,
			&meta.PageNo,
			&meta.SourceURI,
		); err != nil {
			return nil, err
		}
		out[meta.ChunkID] = meta
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ProviderSet is rag data providers.
var ProviderSet = wire.NewSet(NewRAGRepo)

func buildChunkQuery(tenantID string, chunkIDs []string) (string, []any) {
	placeholders := make([]string, 0, len(chunkIDs))
	args := make([]any, 0, len(chunkIDs)+1)
	args = append(args, tenantID)
	for _, id := range chunkIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	query := `SELECT id, kb_id, document_id, document_version_id, content, section, page_no, source_uri
		FROM doc_chunk WHERE tenant_id = ? AND id IN (` + strings.Join(placeholders, ",") + `)`
	return query, args
}
