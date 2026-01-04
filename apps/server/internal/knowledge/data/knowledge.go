package data

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	internaldata "github.com/ZTH7/RAGDesk/apps/server/internal/data"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/google/wire"
)

type knowledgeRepo struct {
	log *log.Helper
	db  *sql.DB

	qdrant     *qdrantClient
	collection string
	storage    objectStorage
}

// NewKnowledgeRepo creates a new knowledge repo.
func NewKnowledgeRepo(data *internaldata.Data, cfg *conf.Data, logger log.Logger) biz.KnowledgeRepo {
	var qc *qdrantClient
	collection := ""
	helper := log.NewHelper(logger)
	if cfg != nil && cfg.Vectordb != nil && cfg.Vectordb.Driver == "qdrant" && cfg.Vectordb.Endpoint != "" {
		qc = newQdrantClient(cfg.Vectordb.Endpoint, cfg.Vectordb.ApiKey)
		collection = cfg.Vectordb.Collection
	}
	if collection == "" {
		collection = "ragdesk_chunks"
	}
	return &knowledgeRepo{
		log:        helper,
		db:         data.DB,
		qdrant:     qc,
		collection: collection,
		storage:    newObjectStorage(cfg, helper),
	}
}

func (r *knowledgeRepo) Ping(ctx context.Context) error {
	if r.db == nil {
		return kerrors.InternalServer("DB_MISSING", "database not initialized")
	}
	return r.db.PingContext(ctx)
}

func (r *knowledgeRepo) CreateKnowledgeBase(ctx context.Context, kb biz.KnowledgeBase) (biz.KnowledgeBase, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.KnowledgeBase{}, err
	}
	if kb.ID == "" {
		kb.ID = uuid.NewString()
	}
	kb.TenantID = tenantID
	if kb.CreatedAt.IsZero() {
		kb.CreatedAt = time.Now()
	}
	if kb.UpdatedAt.IsZero() {
		kb.UpdatedAt = kb.CreatedAt
	}
	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO knowledge_base (id, tenant_id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		kb.ID,
		kb.TenantID,
		kb.Name,
		kb.Description,
		kb.CreatedAt,
		kb.UpdatedAt,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return biz.KnowledgeBase{}, kerrors.Conflict("KB_DUPLICATE", "knowledge base already exists")
		}
		return biz.KnowledgeBase{}, err
	}
	return kb, nil
}

func (r *knowledgeRepo) GetKnowledgeBase(ctx context.Context, id string) (biz.KnowledgeBase, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.KnowledgeBase{}, err
	}
	var kb biz.KnowledgeBase
	err = r.db.QueryRowContext(
		ctx,
		"SELECT id, tenant_id, name, description, created_at, updated_at FROM knowledge_base WHERE tenant_id = ? AND id = ?",
		tenantID,
		id,
	).Scan(&kb.ID, &kb.TenantID, &kb.Name, &kb.Description, &kb.CreatedAt, &kb.UpdatedAt)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return biz.KnowledgeBase{}, kerrors.NotFound("KB_NOT_FOUND", "knowledge base not found")
		}
		return biz.KnowledgeBase{}, err
	}
	return kb, nil
}

func (r *knowledgeRepo) ListKnowledgeBases(ctx context.Context) ([]biz.KnowledgeBase, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, tenant_id, name, description, created_at, updated_at FROM knowledge_base WHERE tenant_id = ? ORDER BY created_at DESC",
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.KnowledgeBase, 0)
	for rows.Next() {
		var kb biz.KnowledgeBase
		if err := rows.Scan(&kb.ID, &kb.TenantID, &kb.Name, &kb.Description, &kb.CreatedAt, &kb.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, kb)
	}
	return items, rows.Err()
}

func (r *knowledgeRepo) CreateDocument(ctx context.Context, doc biz.Document) (biz.Document, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.Document{}, err
	}
	if doc.ID == "" {
		doc.ID = uuid.NewString()
	}
	doc.TenantID = tenantID
	if doc.Status == "" {
		doc.Status = biz.DocumentStatusUploaded
	}
	now := time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = now
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO document (id, tenant_id, kb_id, title, source_type, status, current_version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.ID,
		doc.TenantID,
		doc.KBID,
		doc.Title,
		doc.SourceType,
		doc.Status,
		doc.CurrentVersion,
		doc.CreatedAt,
		doc.UpdatedAt,
	)
	if err != nil {
		return biz.Document{}, err
	}
	return doc, nil
}

func (r *knowledgeRepo) GetDocument(ctx context.Context, id string) (biz.Document, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.Document{}, err
	}
	var doc biz.Document
	err = r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, kb_id, title, source_type, status, current_version, created_at, updated_at
		FROM document WHERE tenant_id = ? AND id = ?`,
		tenantID,
		id,
	).Scan(
		&doc.ID,
		&doc.TenantID,
		&doc.KBID,
		&doc.Title,
		&doc.SourceType,
		&doc.Status,
		&doc.CurrentVersion,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return biz.Document{}, kerrors.NotFound("DOC_NOT_FOUND", "document not found")
		}
		return biz.Document{}, err
	}
	return doc, nil
}

func (r *knowledgeRepo) UpdateDocumentIndexState(ctx context.Context, documentID string, status string, currentVersion int32) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(
		ctx,
		"UPDATE document SET status = ?, current_version = ?, updated_at = ? WHERE tenant_id = ? AND id = ?",
		status,
		currentVersion,
		time.Now(),
		tenantID,
		documentID,
	)
	return err
}

func (r *knowledgeRepo) CreateDocumentVersion(ctx context.Context, v biz.DocumentVersion) (biz.DocumentVersion, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.DocumentVersion{}, err
	}
	if v.ID == "" {
		v.ID = uuid.NewString()
	}
	v.TenantID = tenantID
	if v.Status == "" {
		v.Status = biz.DocumentVersionStatusProcessing
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	rawURI := strings.TrimSpace(v.RawURI)
	if rawURI == "" && v.RawText != "" && r.storage != nil {
		key := fmt.Sprintf("tenant/%s/document/%s/version/%d.txt", tenantID, v.DocumentID, v.Version)
		if uri, err := r.storage.Put(ctx, key, []byte(v.RawText)); err == nil {
			rawURI = uri
		} else if r.log != nil {
			r.log.Warnf("object storage put failed: %v", err)
		}
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO document_version (id, tenant_id, document_id, version, raw_text, raw_uri, status, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID,
		v.TenantID,
		v.DocumentID,
		v.Version,
		v.RawText,
		rawURI,
		v.Status,
		v.ErrorReason,
		v.CreatedAt,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return biz.DocumentVersion{}, kerrors.Conflict("DOC_VERSION_DUPLICATE", "document version already exists")
		}
		return biz.DocumentVersion{}, err
	}
	v.RawURI = rawURI
	return v, nil
}

func (r *knowledgeRepo) GetDocumentVersion(ctx context.Context, id string) (biz.DocumentVersion, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.DocumentVersion{}, err
	}
	var v biz.DocumentVersion
	err = r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, document_id, version, raw_text, raw_uri, status, error_message, created_at
		FROM document_version WHERE tenant_id = ? AND id = ?`,
		tenantID,
		id,
	).Scan(&v.ID, &v.TenantID, &v.DocumentID, &v.Version, &v.RawText, &v.RawURI, &v.Status, &v.ErrorReason, &v.CreatedAt)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return biz.DocumentVersion{}, kerrors.NotFound("DOC_VERSION_NOT_FOUND", "document version not found")
		}
		return biz.DocumentVersion{}, err
	}
	return v, nil
}

func (r *knowledgeRepo) GetDocumentVersionByNumber(ctx context.Context, documentID string, version int32) (biz.DocumentVersion, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.DocumentVersion{}, err
	}
	var v biz.DocumentVersion
	err = r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, document_id, version, raw_text, raw_uri, status, error_message, created_at
		FROM document_version WHERE tenant_id = ? AND document_id = ? AND version = ?`,
		tenantID,
		documentID,
		version,
	).Scan(&v.ID, &v.TenantID, &v.DocumentID, &v.Version, &v.RawText, &v.RawURI, &v.Status, &v.ErrorReason, &v.CreatedAt)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return biz.DocumentVersion{}, kerrors.NotFound("DOC_VERSION_NOT_FOUND", "document version not found")
		}
		return biz.DocumentVersion{}, err
	}
	return v, nil
}

func (r *knowledgeRepo) ListDocumentVersions(ctx context.Context, documentID string) ([]biz.DocumentVersion, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, document_id, version, status, created_at
		FROM document_version WHERE tenant_id = ? AND document_id = ? ORDER BY version DESC`,
		tenantID,
		documentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.DocumentVersion, 0)
	for rows.Next() {
		var v biz.DocumentVersion
		if err := rows.Scan(&v.ID, &v.TenantID, &v.DocumentID, &v.Version, &v.Status, &v.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}

func (r *knowledgeRepo) UpdateDocumentVersionStatus(ctx context.Context, versionID string, status string, errorReason string) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(
		ctx,
		"UPDATE document_version SET status = ?, error_message = ? WHERE tenant_id = ? AND id = ?",
		status,
		errorReason,
		tenantID,
		versionID,
	)
	return err
}

func (r *knowledgeRepo) IndexDocumentVersion(ctx context.Context, req biz.IndexDocumentVersionRequest) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	if r.qdrant == nil {
		return kerrors.InternalServer("VECTORDB_MISSING", "vectordb not configured")
	}
	if r.collection == "" {
		return kerrors.InternalServer("VECTORDB_COLLECTION_MISSING", "vectordb collection missing")
	}
	if req.EmbeddingDim <= 0 {
		return kerrors.InternalServer("EMBEDDING_DIM_INVALID", "embedding dim invalid")
	}
	if err := r.qdrant.EnsureCollection(ctx, r.collection, req.EmbeddingDim); err != nil {
		return err
	}

	points := make([]qdrantPoint, 0, len(req.Chunks))
	now := time.Now()
	for _, item := range req.Chunks {
		ch := item.Chunk
		payload := map[string]any{
			"tenant_id":           tenantID,
			"kb_id":               req.KBID,
			"document_id":         req.DocumentID,
			"document_version_id": req.DocumentVersionID,
			"chunk_id":            ch.ID,
			"chunk_index":         ch.ChunkIndex,
			"token_count":         ch.TokenCount,
			"content_hash":        ch.ContentHash,
			"language":            ch.Language,
			"created_at":          ch.CreatedAt.UnixMilli(),
		}
		points = append(points, qdrantPoint{
			ID:      ch.ID,
			Vector:  item.Vector,
			Payload: payload,
		})
	}
	if err := r.qdrant.UpsertPoints(ctx, r.collection, points); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, item := range req.Chunks {
		ch := item.Chunk
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO doc_chunk
				(id, tenant_id, kb_id, document_id, document_version_id, chunk_index, content, token_count, content_hash, language, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				content = VALUES(content),
				token_count = VALUES(token_count),
				content_hash = VALUES(content_hash),
				language = VALUES(language)`,
			ch.ID,
			tenantID,
			req.KBID,
			req.DocumentID,
			req.DocumentVersionID,
			ch.ChunkIndex,
			ch.Content,
			ch.TokenCount,
			ch.ContentHash,
			ch.Language,
			ch.CreatedAt,
		)
		if err != nil {
			return err
		}

		embeddingID := deterministicEmbeddingID(ch.ID, req.EmbeddingModel)
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO embedding (id, tenant_id, chunk_id, model, created_at)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE model = model`,
			embeddingID,
			tenantID,
			ch.ID,
			req.EmbeddingModel,
			now,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *knowledgeRepo) RollbackDocument(ctx context.Context, documentID string, version int32) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(
		ctx,
		"UPDATE document SET current_version = ?, status = ?, updated_at = ? WHERE tenant_id = ? AND id = ?",
		version,
		biz.DocumentStatusReady,
		time.Now(),
		tenantID,
		documentID,
	)
	return err
}

// ProviderSet is knowledge data providers.
var ProviderSet = wire.NewSet(NewKnowledgeRepo)

func deterministicEmbeddingID(chunkID string, model string) string {
	// Deterministic IDs make ingestion idempotent.
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(chunkID+"|"+model)).String()
}
