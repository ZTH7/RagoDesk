package biz

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// Permission codes (tenant scope).
const (
	PermissionKnowledgeBaseRead  = "tenant.knowledge_base.read"
	PermissionKnowledgeBaseWrite = "tenant.knowledge_base.write"
	PermissionDocumentUpload     = "tenant.document.upload"
	PermissionDocumentRead       = "tenant.document.read"
	PermissionDocumentDelete     = "tenant.document.delete"
	PermissionDocumentReindex    = "tenant.document.reindex"
	PermissionDocumentRollback   = "tenant.document.rollback"
	PermissionBotRead            = "tenant.bot.read"
	PermissionBotKBBind          = "tenant.bot_kb.bind"
	PermissionBotKBUnbind        = "tenant.bot_kb.unbind"
)

// Status constants (MVP).
const (
	DocumentStatusUploaded   = "uploaded"
	DocumentStatusProcessing = "processing"
	DocumentStatusReady      = "ready"
	DocumentStatusFailed     = "failed"

	DocumentVersionStatusProcessing = "processing"
	DocumentVersionStatusReady      = "ready"
	DocumentVersionStatusFailed     = "failed"
)

// KnowledgeBase is a tenant-scoped knowledge base.
type KnowledgeBase struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BotKnowledgeBase links a bot to a knowledge base.
type BotKnowledgeBase struct {
	ID        string
	TenantID  string
	BotID     string
	KBID      string
	Priority  int32
	Weight    float64
	CreatedAt time.Time
}

// Document is a document within a knowledge base.
type Document struct {
	ID             string
	TenantID       string
	KBID           string
	Title          string
	SourceType     string
	Status         string
	CurrentVersion int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// DocumentVersion represents a versioned document content.
type DocumentVersion struct {
	ID          string
	TenantID    string
	DocumentID  string
	Version     int32
	RawText     string
	RawURI      string
	Status      string
	ErrorReason string
	CreatedAt   time.Time
}

// IngestionJob describes a document ingestion task.
type IngestionJob struct {
	TenantID          string
	KBID              string
	DocumentID        string
	DocumentVersionID string
	FallbackVersion   int32
}

// IngestionQueue enqueues ingestion jobs and consumes them.
type IngestionQueue interface {
	Enqueue(ctx context.Context, job IngestionJob) error
	Start(ctx context.Context, handler func(context.Context, IngestionJob) error) error
}

// DocChunk is a chunk of a document version.
type DocChunk struct {
	ID          string
	ChunkIndex  int32
	Content     string
	TokenCount  int32
	ContentHash string
	Language    string
	CreatedAt   time.Time
}

// EmbeddedChunk is a chunk plus its embedding vector.
type EmbeddedChunk struct {
	Chunk  DocChunk
	Vector []float32
}

// IndexDocumentVersionRequest describes indexing input for a document version.
type IndexDocumentVersionRequest struct {
	KBID              string
	DocumentID        string
	DocumentVersionID string
	DocumentTitle     string
	SourceType        string
	EmbeddingModel    string
	EmbeddingDim      int
	Chunks            []EmbeddedChunk
}

// KnowledgeRepo persists knowledge entities and writes to vector store.
type KnowledgeRepo interface {
	Ping(context.Context) error

	CreateKnowledgeBase(ctx context.Context, kb KnowledgeBase) (KnowledgeBase, error)
	GetKnowledgeBase(ctx context.Context, id string) (KnowledgeBase, error)
	ListKnowledgeBases(ctx context.Context) ([]KnowledgeBase, error)
	UpdateKnowledgeBase(ctx context.Context, kb KnowledgeBase) (KnowledgeBase, error)
	DeleteKnowledgeBase(ctx context.Context, id string) error

	CreateDocument(ctx context.Context, doc Document) (Document, error)
	GetDocument(ctx context.Context, id string) (Document, error)
	UpdateDocumentIndexState(ctx context.Context, documentID string, status string, currentVersion int32) error

	CreateDocumentVersion(ctx context.Context, v DocumentVersion) (DocumentVersion, error)
	GetDocumentVersion(ctx context.Context, id string) (DocumentVersion, error)
	GetDocumentVersionByNumber(ctx context.Context, documentID string, version int32) (DocumentVersion, error)
	ListDocumentVersions(ctx context.Context, documentID string) ([]DocumentVersion, error)
	UpdateDocumentVersionStatus(ctx context.Context, versionID string, status string, errorReason string) error

	IndexDocumentVersion(ctx context.Context, req IndexDocumentVersionRequest) error
	RollbackDocument(ctx context.Context, documentID string, version int32) error

	ListDocuments(ctx context.Context, kbID string, limit int, offset int) ([]Document, error)
	DeleteDocument(ctx context.Context, documentID string) error

	ListBotKnowledgeBases(ctx context.Context, botID string) ([]BotKnowledgeBase, error)
	BindBotKnowledgeBase(ctx context.Context, link BotKnowledgeBase) (BotKnowledgeBase, error)
	UnbindBotKnowledgeBase(ctx context.Context, botID string, kbID string) error
}

// KnowledgeUsecase handles knowledge business logic.
type KnowledgeUsecase struct {
	repo KnowledgeRepo
	log  *log.Helper

	queue        IngestionQueue
	asyncEnabled bool

	embedder           EmbeddingProvider
	chunkSizeTokens    int
	chunkOverlapTokens int
	embeddingBatchSize int
}

// NewKnowledgeUsecase creates a new KnowledgeUsecase
func NewKnowledgeUsecase(repo KnowledgeRepo, queue IngestionQueue, logger log.Logger) *KnowledgeUsecase {
	opts := loadIngestionOptions()
	embedder := newEmbeddingProvider(opts)
	uc := &KnowledgeUsecase{
		repo:               repo,
		queue:              queue,
		log:                log.NewHelper(logger),
		embedder:           embedder,
		chunkSizeTokens:    opts.chunkSizeTokens,
		chunkOverlapTokens: opts.chunkOverlapTokens,
		embeddingBatchSize: opts.embeddingBatchSize,
	}
	uc.asyncEnabled = asyncEnabled(queue)
	return uc
}

// StartIngestionConsumer starts consuming ingestion jobs from the queue.
func (uc *KnowledgeUsecase) StartIngestionConsumer(ctx context.Context) error {
	if uc.queue == nil {
		return errors.InternalServer("INGESTION_QUEUE_MISSING", "ingestion queue missing")
	}
	if err := uc.queue.Start(ctx, uc.processIngestion); err != nil {
		return err
	}
	return nil
}

func (uc *KnowledgeUsecase) CreateKnowledgeBase(ctx context.Context, kb KnowledgeBase) (KnowledgeBase, error) {
	kb.Name = strings.TrimSpace(kb.Name)
	if kb.Name == "" {
		return KnowledgeBase{}, errors.BadRequest("KB_NAME_MISSING", "knowledge base name missing")
	}
	return uc.repo.CreateKnowledgeBase(ctx, kb)
}

func (uc *KnowledgeUsecase) GetKnowledgeBase(ctx context.Context, id string) (KnowledgeBase, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return KnowledgeBase{}, errors.BadRequest("KB_ID_MISSING", "knowledge base id missing")
	}
	return uc.repo.GetKnowledgeBase(ctx, id)
}

func (uc *KnowledgeUsecase) ListKnowledgeBases(ctx context.Context) ([]KnowledgeBase, error) {
	return uc.repo.ListKnowledgeBases(ctx)
}

func (uc *KnowledgeUsecase) ListDocuments(ctx context.Context, kbID string, limit int32, offset int32) ([]Document, error) {
	kbID = strings.TrimSpace(kbID)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return uc.repo.ListDocuments(ctx, kbID, int(limit), int(offset))
}

func (uc *KnowledgeUsecase) ListBotKnowledgeBases(ctx context.Context, botID string) ([]BotKnowledgeBase, error) {
	botID = strings.TrimSpace(botID)
	if botID == "" {
		return nil, errors.BadRequest("BOT_ID_MISSING", "bot id missing")
	}
	return uc.repo.ListBotKnowledgeBases(ctx, botID)
}

func (uc *KnowledgeUsecase) UpdateKnowledgeBase(ctx context.Context, kb KnowledgeBase) (KnowledgeBase, error) {
	kb.ID = strings.TrimSpace(kb.ID)
	kb.Name = strings.TrimSpace(kb.Name)
	if kb.ID == "" {
		return KnowledgeBase{}, errors.BadRequest("KB_ID_MISSING", "knowledge base id missing")
	}
	if kb.Name == "" && strings.TrimSpace(kb.Description) == "" {
		return KnowledgeBase{}, errors.BadRequest("KB_UPDATE_EMPTY", "knowledge base update empty")
	}
	return uc.repo.UpdateKnowledgeBase(ctx, kb)
}

func (uc *KnowledgeUsecase) DeleteKnowledgeBase(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.BadRequest("KB_ID_MISSING", "knowledge base id missing")
	}
	return uc.repo.DeleteKnowledgeBase(ctx, id)
}

func (uc *KnowledgeUsecase) DeleteDocument(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.BadRequest("DOC_ID_MISSING", "document id missing")
	}
	return uc.repo.DeleteDocument(ctx, id)
}

func (uc *KnowledgeUsecase) BindBotKnowledgeBase(ctx context.Context, botID, kbID string, priority int32, weight float64) (BotKnowledgeBase, error) {
	botID = strings.TrimSpace(botID)
	kbID = strings.TrimSpace(kbID)
	if botID == "" {
		return BotKnowledgeBase{}, errors.BadRequest("BOT_ID_MISSING", "bot id missing")
	}
	if kbID == "" {
		return BotKnowledgeBase{}, errors.BadRequest("KB_ID_MISSING", "knowledge base id missing")
	}
	if weight <= 0 {
		weight = 1
	}
	if _, err := uc.repo.GetKnowledgeBase(ctx, kbID); err != nil {
		return BotKnowledgeBase{}, err
	}
	return uc.repo.BindBotKnowledgeBase(ctx, BotKnowledgeBase{
		BotID:    botID,
		KBID:     kbID,
		Priority: priority,
		Weight:   weight,
	})
}

func (uc *KnowledgeUsecase) UnbindBotKnowledgeBase(ctx context.Context, botID, kbID string) error {
	botID = strings.TrimSpace(botID)
	kbID = strings.TrimSpace(kbID)
	if botID == "" {
		return errors.BadRequest("BOT_ID_MISSING", "bot id missing")
	}
	if kbID == "" {
		return errors.BadRequest("KB_ID_MISSING", "knowledge base id missing")
	}
	return uc.repo.UnbindBotKnowledgeBase(ctx, botID, kbID)
}

func (uc *KnowledgeUsecase) UploadDocument(ctx context.Context, kbID, title, sourceType, content string) (Document, DocumentVersion, error) {
	kbID = strings.TrimSpace(kbID)
	title = strings.TrimSpace(title)
	sourceType = normalizeSourceType(sourceType)
	content = strings.TrimSpace(content)
	if kbID == "" {
		return Document{}, DocumentVersion{}, errors.BadRequest("KB_ID_MISSING", "kb_id missing")
	}
	if title == "" {
		return Document{}, DocumentVersion{}, errors.BadRequest("DOC_TITLE_MISSING", "title missing")
	}
	if content == "" {
		return Document{}, DocumentVersion{}, errors.BadRequest("DOC_CONTENT_MISSING", "content missing")
	}
	// Ensure KB exists (tenant scoped).
	if _, err := uc.repo.GetKnowledgeBase(ctx, kbID); err != nil {
		return Document{}, DocumentVersion{}, err
	}
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return Document{}, DocumentVersion{}, err
	}

	doc, err := uc.repo.CreateDocument(ctx, Document{
		KBID:           kbID,
		Title:          title,
		SourceType:     sourceType,
		Status:         DocumentStatusProcessing,
		CurrentVersion: 0,
	})
	if err != nil {
		return Document{}, DocumentVersion{}, err
	}
	ver, err := uc.repo.CreateDocumentVersion(ctx, DocumentVersion{
		DocumentID: doc.ID,
		Version:    1,
		RawText:    content,
		Status:     DocumentVersionStatusProcessing,
	})
	if err != nil {
		return Document{}, DocumentVersion{}, err
	}

	job := IngestionJob{
		TenantID:          tenantID,
		KBID:              kbID,
		DocumentID:        doc.ID,
		DocumentVersionID: ver.ID,
		FallbackVersion:   0,
	}
	if uc.enqueueIngestion(ctx, job) {
		return doc, ver, nil
	}
	if err := uc.processIngestion(ctx, job); err != nil {
		return Document{}, DocumentVersion{}, err
	}
	doc.Status = DocumentStatusReady
	doc.CurrentVersion = 1
	ver.Status = DocumentVersionStatusReady
	return doc, ver, nil
}

func (uc *KnowledgeUsecase) GetDocument(ctx context.Context, id string) (Document, []DocumentVersion, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Document{}, nil, errors.BadRequest("DOC_ID_MISSING", "document id missing")
	}
	doc, err := uc.repo.GetDocument(ctx, id)
	if err != nil {
		return Document{}, nil, err
	}
	versions, err := uc.repo.ListDocumentVersions(ctx, id)
	if err != nil {
		return Document{}, nil, err
	}
	return doc, versions, nil
}

func (uc *KnowledgeUsecase) ReindexDocument(ctx context.Context, id string) (DocumentVersion, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return DocumentVersion{}, errors.BadRequest("DOC_ID_MISSING", "document id missing")
	}
	doc, err := uc.repo.GetDocument(ctx, id)
	if err != nil {
		return DocumentVersion{}, err
	}
	if doc.CurrentVersion <= 0 {
		return DocumentVersion{}, errors.New(412, "DOC_VERSION_MISSING", "document has no version")
	}
	current, err := uc.repo.GetDocumentVersionByNumber(ctx, id, doc.CurrentVersion)
	if err != nil {
		return DocumentVersion{}, err
	}
	nextVersion := doc.CurrentVersion + 1
	ver, err := uc.repo.CreateDocumentVersion(ctx, DocumentVersion{
		DocumentID: id,
		Version:    nextVersion,
		RawText:    current.RawText,
		Status:     DocumentVersionStatusProcessing,
	})
	if err != nil {
		return DocumentVersion{}, err
	}
	_ = uc.repo.UpdateDocumentIndexState(ctx, doc.ID, DocumentStatusProcessing, doc.CurrentVersion)

	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return DocumentVersion{}, err
	}
	job := IngestionJob{
		TenantID:          tenantID,
		KBID:              doc.KBID,
		DocumentID:        doc.ID,
		DocumentVersionID: ver.ID,
		FallbackVersion:   doc.CurrentVersion,
	}
	if uc.enqueueIngestion(ctx, job) {
		return ver, nil
	}
	if err := uc.processIngestion(ctx, job); err != nil {
		return DocumentVersion{}, err
	}
	ver.Status = DocumentVersionStatusReady
	return ver, nil
}

func (uc *KnowledgeUsecase) RollbackDocument(ctx context.Context, id string, version int32) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.BadRequest("DOC_ID_MISSING", "document id missing")
	}
	if version <= 0 {
		return errors.BadRequest("DOC_VERSION_INVALID", "invalid version")
	}
	v, err := uc.repo.GetDocumentVersionByNumber(ctx, id, version)
	if err != nil {
		return err
	}
	if v.Status != DocumentVersionStatusReady {
		return errors.New(412, "DOC_VERSION_NOT_READY", "target version not ready")
	}
	return uc.repo.RollbackDocument(ctx, id, version)
}

func (uc *KnowledgeUsecase) enqueueIngestion(ctx context.Context, job IngestionJob) bool {
	if !uc.asyncEnabled || uc.queue == nil {
		return false
	}
	if err := uc.queue.Enqueue(ctx, job); err != nil {
		uc.log.Warnf("enqueue ingestion failed: %v", err)
		return false
	}
	return true
}

func (uc *KnowledgeUsecase) processIngestion(ctx context.Context, job IngestionJob) error {
	if job.TenantID == "" {
		return errors.Forbidden("TENANT_MISSING", "tenant missing")
	}
	ctx = withTenantID(ctx, job.TenantID)
	version, err := uc.repo.GetDocumentVersion(ctx, job.DocumentVersionID)
	if err != nil {
		return err
	}
	doc, err := uc.repo.GetDocument(ctx, job.DocumentID)
	if err != nil {
		return err
	}
	sourceType := normalizeSourceType(doc.SourceType)
	parsed, err := parseContent(ctx, sourceType, version.RawText)
	if err != nil {
		_ = uc.repo.UpdateDocumentVersionStatus(ctx, version.ID, DocumentVersionStatusFailed, err.Error())
		_ = uc.repo.UpdateDocumentIndexState(ctx, job.DocumentID, DocumentStatusFailed, job.FallbackVersion)
		return err
	}
	content := prepareContentNormalized(sourceType, parsed)
	if content == "" {
		return errors.BadRequest("DOC_CONTENT_MISSING", "document content missing")
	}
	chunks := buildChunks(content, version.ID, uc.chunkSizeTokens, uc.chunkOverlapTokens)
	if len(chunks) == 0 {
		return errors.BadRequest("DOC_CHUNKS_EMPTY", "document chunks empty")
	}
	embedded, err := uc.embedChunks(ctx, chunks)
	if err != nil {
		return err
	}
	indexReq := IndexDocumentVersionRequest{
		KBID:              job.KBID,
		DocumentID:        job.DocumentID,
		DocumentVersionID: version.ID,
		DocumentTitle:     doc.Title,
		SourceType:        sourceType,
		EmbeddingModel:    uc.embedder.Model(),
		EmbeddingDim:      uc.embedder.Dim(),
		Chunks:            embedded,
	}
	if err := uc.repo.IndexDocumentVersion(ctx, indexReq); err != nil {
		_ = uc.repo.UpdateDocumentVersionStatus(ctx, version.ID, DocumentVersionStatusFailed, err.Error())
		_ = uc.repo.UpdateDocumentIndexState(ctx, job.DocumentID, DocumentStatusFailed, job.FallbackVersion)
		return err
	}
	_ = uc.repo.UpdateDocumentVersionStatus(ctx, version.ID, DocumentVersionStatusReady, "")
	_ = uc.repo.UpdateDocumentIndexState(ctx, job.DocumentID, DocumentStatusReady, version.Version)
	return nil
}

// ProviderSet is knowledge biz providers.
var ProviderSet = wire.NewSet(NewKnowledgeUsecase)
