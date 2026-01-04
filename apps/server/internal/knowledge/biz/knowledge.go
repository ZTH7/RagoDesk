package biz

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
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

// EmbeddingProvider embeds text into vector representations.
type EmbeddingProvider interface {
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
	Model() string
	Dim() int
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

const (
	defaultChunkSizeTokens    = 400
	defaultChunkOverlapTokens = 50
	defaultEmbeddingModel     = "fake-embedding-v1"
	defaultEmbeddingDim       = 384
	defaultEmbeddingProvider  = "fake"
)

type ingestionOptions struct {
	chunkSizeTokens    int
	chunkOverlapTokens int
	embeddingModel     string
	embeddingDim       int
	embeddingProvider  string
	embeddingEndpoint  string
	embeddingAPIKey    string
	embeddingTimeoutMs int
}

func loadIngestionOptions() ingestionOptions {
	provider := envString("RAGDESK_EMBEDDING_PROVIDER", defaultEmbeddingProvider)
	embeddingDim := defaultEmbeddingDim
	if raw := strings.TrimSpace(os.Getenv("RAGDESK_EMBEDDING_DIM")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			embeddingDim = parsed
		}
	} else if strings.EqualFold(provider, "openai") || strings.EqualFold(provider, "http") {
		embeddingDim = 0
	}
	opts := ingestionOptions{
		chunkSizeTokens:    envInt("RAGDESK_CHUNK_SIZE_TOKENS", defaultChunkSizeTokens),
		chunkOverlapTokens: envInt("RAGDESK_CHUNK_OVERLAP_TOKENS", defaultChunkOverlapTokens),
		embeddingModel:     envString("RAGDESK_EMBEDDING_MODEL", defaultEmbeddingModel),
		embeddingDim:       embeddingDim,
		embeddingProvider:  provider,
		embeddingEndpoint:  envString("RAGDESK_EMBEDDING_ENDPOINT", ""),
		embeddingAPIKey:    envString("RAGDESK_EMBEDDING_API_KEY", ""),
		embeddingTimeoutMs: envInt("RAGDESK_EMBEDDING_TIMEOUT_MS", 15000),
	}
	if opts.chunkSizeTokens <= 0 {
		opts.chunkSizeTokens = defaultChunkSizeTokens
	}
	if opts.chunkOverlapTokens < 0 {
		opts.chunkOverlapTokens = defaultChunkOverlapTokens
	}
	if opts.embeddingDim < 0 {
		opts.embeddingDim = defaultEmbeddingDim
	}
	return opts
}

func newEmbeddingProvider(opts ingestionOptions) EmbeddingProvider {
	switch strings.ToLower(strings.TrimSpace(opts.embeddingProvider)) {
	case "", "fake":
		return fakeEmbeddingProvider{
			model: opts.embeddingModel,
			dim:   opts.embeddingDim,
		}
	case "openai", "http":
		endpoint := strings.TrimSpace(opts.embeddingEndpoint)
		if endpoint == "" {
			return fakeEmbeddingProvider{
				model: opts.embeddingModel,
				dim:   opts.embeddingDim,
			}
		}
		return &openAIEmbeddingProvider{
			endpoint: strings.TrimRight(endpoint, "/"),
			apiKey:   opts.embeddingAPIKey,
			model:    opts.embeddingModel,
			dim:      opts.embeddingDim,
			client: &http.Client{
				Timeout: time.Duration(opts.embeddingTimeoutMs) * time.Millisecond,
			},
		}
	default:
		// fallback to fake provider for now
		return fakeEmbeddingProvider{
			model: opts.embeddingModel,
			dim:   opts.embeddingDim,
		}
	}
}

type fakeEmbeddingProvider struct {
	model string
	dim   int
}

func (p fakeEmbeddingProvider) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	out := make([][]float32, 0, len(inputs))
	for _, text := range inputs {
		out = append(out, deterministicEmbedding(text, p.dim))
	}
	return out, nil
}

func (p fakeEmbeddingProvider) Model() string {
	if p.model == "" {
		return defaultEmbeddingModel
	}
	return p.model
}

func (p fakeEmbeddingProvider) Dim() int {
	if p.dim <= 0 {
		return defaultEmbeddingDim
	}
	return p.dim
}

type openAIEmbeddingProvider struct {
	endpoint string
	apiKey   string
	model    string
	dim      int
	client   *http.Client
}

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func (p *openAIEmbeddingProvider) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	if p == nil || p.endpoint == "" {
		return nil, errors.InternalServer("EMBEDDING_ENDPOINT_MISSING", "embedding endpoint missing")
	}
	if p.client == nil {
		p.client = &http.Client{Timeout: 15 * time.Second}
	}
	reqBody := openAIEmbeddingRequest{
		Model: p.Model(),
		Input: inputs,
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	url := p.endpoint
	if !strings.HasSuffix(url, "/embeddings") {
		url = url + "/embeddings"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.InternalServer("EMBEDDING_REQUEST_FAILED", "embedding request failed")
	}
	var parsed openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	out := make([][]float32, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		vec := make([]float32, 0, len(item.Embedding))
		for _, v := range item.Embedding {
			vec = append(vec, float32(v))
		}
		out = append(out, vec)
	}
	if len(out) > 0 && p.dim <= 0 {
		p.dim = len(out[0])
	}
	return out, nil
}

func (p *openAIEmbeddingProvider) Model() string {
	if p.model == "" {
		return defaultEmbeddingModel
	}
	return p.model
}

func (p *openAIEmbeddingProvider) Dim() int {
	if p.dim <= 0 {
		return defaultEmbeddingDim
	}
	return p.dim
}

func (uc *KnowledgeUsecase) embedChunks(ctx context.Context, chunks []DocChunk) ([]EmbeddedChunk, error) {
	if len(chunks) == 0 {
		return nil, nil
	}
	if uc.embedder == nil {
		uc.embedder = fakeEmbeddingProvider{model: defaultEmbeddingModel, dim: defaultEmbeddingDim}
	}
	texts := make([]string, 0, len(chunks))
	for _, ch := range chunks {
		texts = append(texts, ch.Content)
	}
	vectors, err := uc.embedder.Embed(ctx, texts)
	if err != nil {
		return nil, err
	}
	if len(vectors) != len(chunks) {
		return nil, errors.InternalServer("EMBEDDING_COUNT_MISMATCH", "embedding count mismatch")
	}
	out := make([]EmbeddedChunk, 0, len(chunks))
	for i, ch := range chunks {
		out = append(out, EmbeddedChunk{
			Chunk:  ch,
			Vector: vectors[i],
		})
	}
	return out, nil
}

func buildChunks(rawText, docVersionID string, chunkSize, overlap int) []DocChunk {
	chunkSize = clampInt(chunkSize, 1, 8192)
	overlap = clampInt(overlap, 0, chunkSize-1)
	parts := splitByTokens(rawText, chunkSize, overlap)

	out := make([]DocChunk, 0, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		chunkIndex := int32(i)
		chunkID := deterministicChunkID(docVersionID, chunkIndex)
		createdAt := time.Now()
		chunk := DocChunk{
			ID:          chunkID,
			ChunkIndex:  chunkIndex,
			Content:     part,
			TokenCount:  int32(estimateTokenCount(part)),
			ContentHash: sha256Hex(part),
			Language:    detectLanguage(part),
			CreatedAt:   createdAt,
		}
		out = append(out, chunk)
	}
	return out
}

type tokenSpan struct {
	start int
	end   int
}

func splitByTokens(s string, chunkSize, overlap int) []string {
	runes := []rune(s)
	if len(runes) == 0 {
		return nil
	}
	tokens := tokenizeSpans(runes)
	if len(tokens) == 0 {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = len(tokens)
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize - 1
	}
	step := chunkSize - overlap
	if step <= 0 {
		step = chunkSize
	}
	out := make([]string, 0, (len(tokens)+step-1)/step)
	for start := 0; start < len(tokens); start += step {
		end := start + chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}
		startIdx := tokens[start].start
		endIdx := tokens[end-1].end
		out = append(out, string(runes[startIdx:endIdx]))
		if end == len(tokens) {
			break
		}
	}
	return out
}

func tokenizeSpans(runes []rune) []tokenSpan {
	out := make([]tokenSpan, 0, len(runes)/2+1)
	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case isCJK(r):
			out = append(out, tokenSpan{start: i, end: i + 1})
			i++
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			start := i
			i++
			for i < len(runes) {
				if isCJK(runes[i]) || !(unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i])) {
					break
				}
				i++
			}
			out = append(out, tokenSpan{start: start, end: i})
		default:
			i++
		}
	}
	return out
}

func estimateTokenCount(s string) int {
	return len(tokenizeSpans([]rune(s)))
}

func detectLanguage(s string) string {
	var cjk, latin int
	for _, r := range s {
		switch {
		case isCJK(r):
			cjk++
		case r <= unicode.MaxLatin1 && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			latin++
		}
	}
	if cjk >= 10 && cjk >= latin {
		return "zh"
	}
	if latin >= 10 && latin > cjk {
		return "en"
	}
	return ""
}

func isCJK(r rune) bool {
	// Common CJK Unified Ideographs blocks.
	return (r >= 0x4E00 && r <= 0x9FFF) || (r >= 0x3400 && r <= 0x4DBF) || (r >= 0x20000 && r <= 0x2A6DF)
}

func deterministicChunkID(docVersionID string, chunkIndex int32) string {
	// Deterministic IDs make ingestion idempotent for the same document_version.
	key := fmt.Sprintf("%s:%d", docVersionID, chunkIndex)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(key)).String()
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func deterministicEmbedding(s string, dim int) []float32 {
	if dim <= 0 {
		return nil
	}
	sum := sha256.Sum256([]byte(s))
	seed := binary.LittleEndian.Uint64(sum[:8])
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	out := make([]float32, dim)
	for i := 0; i < dim; i++ {
		// [-1, 1)
		out[i] = float32(rng.Float64()*2.0 - 1.0)
	}
	return out
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func asyncEnabled(queue IngestionQueue) bool {
	if queue == nil {
		return false
	}
	value := strings.TrimSpace(os.Getenv("RAGDESK_INGESTION_ASYNC"))
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func tenantIDFromContext(ctx context.Context) (string, error) {
	if value, ok := tenant.TenantID(ctx); ok {
		return value, nil
	}
	return "", errors.Forbidden("TENANT_MISSING", "tenant missing")
}

func withTenantID(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		return ctx
	}
	return tenant.WithTenantID(ctx, tenantID)
}

func cleanContent(input string) string {
	if input == "" {
		return ""
	}
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		out = append(out, strings.Join(fields, " "))
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func normalizeSourceType(sourceType string) string {
	value := strings.ToLower(strings.TrimSpace(sourceType))
	switch value {
	case "", "text", "plain", "txt":
		return "text"
	case "md", "markdown":
		return "markdown"
	case "html", "htm":
		return "html"
	case "doc", "docx":
		return "doc"
	case "pdf":
		return "pdf"
	case "url", "link":
		return "url"
	default:
		return value
	}
}

func prepareContentNormalized(sourceType, content string) string {
	switch sourceType {
	case "markdown":
		content = stripMarkdown(content)
	case "html":
		content = stripHTMLTags(content)
	}
	return cleanContent(content)
}

const maxDocumentBytes = 5 << 20

func parseContent(ctx context.Context, sourceType, raw string) (string, error) {
	switch sourceType {
	case "url":
		return fetchURLText(ctx, raw)
	case "doc":
		return parseDocxBase64(raw)
	default:
		// pdf / text / markdown / html fall back to raw text
		return raw, nil
	}
}

func fetchURLText(ctx context.Context, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.BadRequest("DOC_URL_EMPTY", "document url missing")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.BadRequest("DOC_URL_INVALID", "document url invalid")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.BadRequest("DOC_URL_FETCH_FAILED", "document url fetch failed")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDocumentBytes))
	if err != nil {
		return "", err
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	text := string(body)
	if strings.Contains(contentType, "text/html") {
		text = stripHTMLTags(text)
	}
	return text, nil
}

func parseDocxBase64(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.BadRequest("DOCX_CONTENT_EMPTY", "docx content missing")
	}
	payload, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", errors.BadRequest("DOCX_BASE64_INVALID", "docx content must be base64")
	}
	readerAt := bytes.NewReader(payload)
	zr, err := zip.NewReader(readerAt, int64(len(payload)))
	if err != nil {
		return "", err
	}
	var xmlFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			xmlFile = f
			break
		}
	}
	if xmlFile == nil {
		return "", errors.BadRequest("DOCX_XML_MISSING", "docx document xml missing")
	}
	rc, err := xmlFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	xmlBytes, err := io.ReadAll(io.LimitReader(rc, maxDocumentBytes))
	if err != nil {
		return "", err
	}
	decoder := xml.NewDecoder(bytes.NewReader(xmlBytes))
	var builder strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "t" {
			continue
		}
		var text string
		if err := decoder.DecodeElement(&text, &start); err != nil {
			return "", err
		}
		if text != "" {
			builder.WriteString(text)
			builder.WriteString(" ")
		}
	}
	return builder.String(), nil
}

func stripMarkdown(input string) string {
	if input == "" {
		return ""
	}
	lines := strings.Split(input, "\n")
	out := make([]string, 0, len(lines))
	inCode := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			out = append(out, line)
			continue
		}
		line = strings.TrimLeft(line, " \t")
		if strings.HasPrefix(line, ">") {
			line = strings.TrimSpace(strings.TrimPrefix(line, ">"))
		}
		if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ") {
			line = strings.TrimSpace(line[2:])
		}
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.ReplaceAll(line, "`", "")
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func stripHTMLTags(input string) string {
	if input == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(input))
	inTag := false
	for _, r := range input {
		switch r {
		case '<':
			inTag = true
			b.WriteRune(' ')
		case '>':
			inTag = false
			b.WriteRune(' ')
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// ProviderSet is knowledge biz providers.
var ProviderSet = wire.NewSet(NewKnowledgeUsecase)
