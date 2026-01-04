package biz

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
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
}

// NewKnowledgeUsecase creates a new KnowledgeUsecase
func NewKnowledgeUsecase(repo KnowledgeRepo, queue IngestionQueue, logger log.Logger) *KnowledgeUsecase {
	uc := &KnowledgeUsecase{
		repo:  repo,
		queue: queue,
		log:   log.NewHelper(logger),
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
	content := prepareContentNormalized(doc.SourceType, version.RawText)
	if content == "" {
		return errors.BadRequest("DOC_CONTENT_MISSING", "document content missing")
	}
	embedded := embedChunks(content, version.ID, defaultChunkSizeRunes, defaultChunkOverlapRunes, defaultEmbeddingDim)
	indexReq := IndexDocumentVersionRequest{
		KBID:              job.KBID,
		DocumentID:        job.DocumentID,
		DocumentVersionID: version.ID,
		EmbeddingModel:    defaultEmbeddingModel,
		EmbeddingDim:      defaultEmbeddingDim,
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
	defaultChunkSizeRunes    = 800
	defaultChunkOverlapRunes = 100
	defaultEmbeddingModel    = "fake-embedding-v1"
	defaultEmbeddingDim      = 384
)

func embedChunks(rawText, docVersionID string, chunkSize, overlap, embeddingDim int) []EmbeddedChunk {
	chunkSize = clampInt(chunkSize, 1, 8192)
	overlap = clampInt(overlap, 0, chunkSize-1)
	parts := splitByRunes(rawText, chunkSize, overlap)

	out := make([]EmbeddedChunk, 0, len(parts))
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
			TokenCount:  int32(runeTokenCount(part)),
			ContentHash: sha256Hex(part),
			Language:    detectLanguage(part),
			CreatedAt:   createdAt,
		}
		out = append(out, EmbeddedChunk{
			Chunk:  chunk,
			Vector: deterministicEmbedding(part, embeddingDim),
		})
	}
	return out
}

func splitByRunes(s string, chunkSize, overlap int) []string {
	runes := []rune(s)
	if len(runes) == 0 {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = len(runes)
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
	out := make([]string, 0, (len(runes)+step-1)/step)
	for start := 0; start < len(runes); start += step {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
		if end == len(runes) {
			break
		}
	}
	return out
}

func runeTokenCount(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		count++
	}
	return count
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
