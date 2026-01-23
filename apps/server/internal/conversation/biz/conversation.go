package biz

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/ZTH7/RAGDesk/apps/server/internal/paging"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/google/wire"
)

const (
	SessionStatusBot    = "bot"
	SessionStatusClosed = "closed"

	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"

	EventOpen       = "open"
	EventClose      = "close"
	EventRefusal    = "refusal"
	EventEscalation = "escalation"
)

const (
	defaultRetentionDays        = 0
	defaultPurgeIntervalMinutes = 60
)

// Session represents a chat session.
type Session struct {
	ID           string
	TenantID     string
	BotID        string
	Status       string
	CloseReason  string
	UserExternal string
	Metadata     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ClosedAt     time.Time
}

// Message represents a chat message.
type Message struct {
	ID         string
	TenantID   string
	SessionID  string
	Role       string
	Content    string
	Confidence float32
	References string
	CreatedAt  time.Time
}

// SessionEvent represents a session audit event.
type SessionEvent struct {
	ID        string
	TenantID  string
	SessionID string
	EventType string
	Detail    string
	CreatedAt time.Time
}

// MessageFeedback represents user feedback for a message.
type MessageFeedback struct {
	ID         string
	TenantID   string
	SessionID  string
	MessageID  string
	Rating     int32
	Comment    string
	Correction string
	CreatedAt  time.Time
}

// ConversationRepo is a repository interface.
type ConversationRepo interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	GetSession(ctx context.Context, sessionID string) (Session, error)
	ListSessions(ctx context.Context, botID string, limit int, offset int) ([]Session, error)
	CloseSession(ctx context.Context, sessionID string, closeReason string, closedAt time.Time) error
	PurgeExpired(ctx context.Context, cutoff time.Time) error

	CreateMessages(ctx context.Context, messages []Message) error
	ListMessages(ctx context.Context, sessionID string, limit int, offset int) ([]Message, error)

	CreateEvent(ctx context.Context, event SessionEvent) error
	CreateFeedback(ctx context.Context, feedback MessageFeedback) error
}

// ConversationUsecase handles conversation business logic.
type ConversationUsecase struct {
	repo          ConversationRepo
	retentionDays int
	purgeInterval time.Duration
	lastPurge     time.Time
	mu            sync.Mutex
}

// NewConversationUsecase creates a new ConversationUsecase
func NewConversationUsecase(repo ConversationRepo, cfg *conf.Data) *ConversationUsecase {
	retentionDays, purgeInterval := loadRetentionPolicy(cfg)
	return &ConversationUsecase{
		repo:          repo,
		retentionDays: retentionDays,
		purgeInterval: purgeInterval,
	}
}

func (uc *ConversationUsecase) CreateSession(ctx context.Context, botID string, userExternal string, metadata map[string]any) (Session, error) {
	botID = strings.TrimSpace(botID)
	if botID == "" {
		return Session{}, errors.BadRequest("BOT_ID_MISSING", "bot id missing")
	}
	uc.maybePurge(ctx)
	metaJSON := ""
	if metadata != nil {
		if raw, err := json.Marshal(metadata); err == nil {
			metaJSON = string(raw)
		}
	}
	now := time.Now()
	session := Session{
		ID:           uuid.NewString(),
		BotID:        botID,
		Status:       SessionStatusBot,
		UserExternal: strings.TrimSpace(userExternal),
		Metadata:     metaJSON,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	created, err := uc.repo.CreateSession(ctx, session)
	if err != nil {
		return Session{}, err
	}
	_ = uc.repo.CreateEvent(ctx, SessionEvent{
		ID:        uuid.NewString(),
		SessionID: created.ID,
		EventType: EventOpen,
		CreatedAt: now,
	})
	return created, nil
}

func (uc *ConversationUsecase) GetSession(ctx context.Context, sessionID string, includeMessages bool, limit int, offset int) (Session, []Message, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return Session{}, nil, errors.BadRequest("SESSION_ID_MISSING", "session id missing")
	}
	uc.maybePurge(ctx)
	session, err := uc.repo.GetSession(ctx, sessionID)
	if err != nil {
		return Session{}, nil, err
	}
	if !includeMessages {
		return session, nil, nil
	}
	limit, offset = paging.Normalize(limit, offset)
	messages, err := uc.repo.ListMessages(ctx, sessionID, limit, offset)
	if err != nil {
		return Session{}, nil, err
	}
	return session, messages, nil
}

func (uc *ConversationUsecase) CloseSession(ctx context.Context, sessionID string, closeReason string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return errors.BadRequest("SESSION_ID_MISSING", "session id missing")
	}
	uc.maybePurge(ctx)
	session, err := uc.repo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session.Status == SessionStatusClosed {
		return nil
	}
	if !canCloseSession(session.Status) {
		return errors.New(412, "SESSION_STATUS_INVALID", "session status cannot be closed")
	}
	now := time.Now()
	if err := uc.repo.CloseSession(ctx, sessionID, strings.TrimSpace(closeReason), now); err != nil {
		return err
	}
	_ = uc.repo.CreateEvent(ctx, SessionEvent{
		ID:        uuid.NewString(),
		SessionID: sessionID,
		EventType: EventClose,
		Detail:    strings.TrimSpace(closeReason),
		CreatedAt: now,
	})
	if isEscalationReason(closeReason) {
		_ = uc.repo.CreateEvent(ctx, SessionEvent{
			ID:        uuid.NewString(),
			SessionID: sessionID,
			EventType: EventEscalation,
			Detail:    strings.TrimSpace(closeReason),
			CreatedAt: now,
		})
	}
	return nil
}

func (uc *ConversationUsecase) ListSessions(ctx context.Context, botID string, limit int, offset int) ([]Session, error) {
	uc.maybePurge(ctx)
	limit, offset = paging.Normalize(limit, offset)
	return uc.repo.ListSessions(ctx, botID, limit, offset)
}

func (uc *ConversationUsecase) ListMessages(ctx context.Context, sessionID string, limit int, offset int) ([]Message, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, errors.BadRequest("SESSION_ID_MISSING", "session id missing")
	}
	uc.maybePurge(ctx)
	limit, offset = paging.Normalize(limit, offset)
	return uc.repo.ListMessages(ctx, sessionID, limit, offset)
}

func (uc *ConversationUsecase) CreateFeedback(ctx context.Context, sessionID string, messageID string, rating int32, comment string, correction string) error {
	sessionID = strings.TrimSpace(sessionID)
	messageID = strings.TrimSpace(messageID)
	if sessionID == "" || messageID == "" {
		return errors.BadRequest("FEEDBACK_INVALID", "session_id or message_id missing")
	}
	uc.maybePurge(ctx)
	if rating == 0 {
		return errors.BadRequest("FEEDBACK_RATING_INVALID", "rating missing")
	}
	return uc.repo.CreateFeedback(ctx, MessageFeedback{
		ID:         uuid.NewString(),
		SessionID:  sessionID,
		MessageID:  messageID,
		Rating:     rating,
		Comment:    strings.TrimSpace(comment),
		Correction: strings.TrimSpace(correction),
		CreatedAt:  time.Now(),
	})
}

func (uc *ConversationUsecase) RecordRAGExchange(ctx context.Context, sessionID string, botID string, userMessage string, answer string, confidence float32, refused bool, referencesJSON string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", nil
	}
	uc.maybePurge(ctx)
	session, err := uc.repo.GetSession(ctx, sessionID)
	if err != nil {
		return "", err
	}
	if botID != "" && session.BotID != "" && botID != session.BotID {
		return "", errors.BadRequest("SESSION_BOT_MISMATCH", "session bot mismatch")
	}
	if session.Status == SessionStatusClosed {
		return "", errors.New(412, "SESSION_CLOSED", "session already closed")
	}
	now := time.Now()
	userID := uuid.NewString()
	user := Message{
		ID:        userID,
		SessionID: sessionID,
		Role:      MessageRoleUser,
		Content:   strings.TrimSpace(userMessage),
		CreatedAt: now,
	}
	assistant := Message{
		ID:         uuid.NewString(),
		SessionID:  sessionID,
		Role:       MessageRoleAssistant,
		Content:    strings.TrimSpace(answer),
		Confidence: confidence,
		References: strings.TrimSpace(referencesJSON),
		CreatedAt:  now,
	}
	if err := uc.repo.CreateMessages(ctx, []Message{user, assistant}); err != nil {
		return "", err
	}
	if refused {
		_ = uc.repo.CreateEvent(ctx, SessionEvent{
			ID:        uuid.NewString(),
			SessionID: sessionID,
			EventType: EventRefusal,
			CreatedAt: now,
		})
	}
	return userID, nil
}

func canCloseSession(status string) bool {
	switch status {
	case SessionStatusBot:
		return true
	case SessionStatusClosed:
		return false
	default:
		return false
	}
}

func loadRetentionPolicy(cfg *conf.Data) (int, time.Duration) {
	retentionDays := defaultRetentionDays
	purgeMinutes := defaultPurgeIntervalMinutes
	if cfg != nil && cfg.Conversation != nil {
		if cfg.Conversation.RetentionDays > 0 {
			retentionDays = int(cfg.Conversation.RetentionDays)
		}
		if cfg.Conversation.PurgeIntervalMinutes > 0 {
			purgeMinutes = int(cfg.Conversation.PurgeIntervalMinutes)
		}
	}
	if retentionDays < 0 {
		retentionDays = 0
	}
	if purgeMinutes <= 0 {
		purgeMinutes = defaultPurgeIntervalMinutes
	}
	return retentionDays, time.Duration(purgeMinutes) * time.Minute
}

func (uc *ConversationUsecase) maybePurge(ctx context.Context) {
	if uc == nil || uc.retentionDays <= 0 || uc.purgeInterval <= 0 {
		return
	}
	tenantID, ok := tenant.TenantID(ctx)
	if !ok || strings.TrimSpace(tenantID) == "" {
		return
	}
	now := time.Now()
	uc.mu.Lock()
	if !uc.lastPurge.IsZero() && now.Sub(uc.lastPurge) < uc.purgeInterval {
		uc.mu.Unlock()
		return
	}
	uc.lastPurge = now
	uc.mu.Unlock()
	cutoff := now.AddDate(0, 0, -uc.retentionDays)
	purgeCtx := tenant.WithTenantID(context.Background(), tenantID)
	go func() {
		_ = uc.repo.PurgeExpired(purgeCtx, cutoff)
	}()
}

func isEscalationReason(reason string) bool {
	reason = strings.ToLower(strings.TrimSpace(reason))
	return strings.Contains(reason, "escalat") || strings.Contains(reason, "handoff")
}

// ProviderSet is conversation biz providers.
var ProviderSet = wire.NewSet(NewConversationUsecase)
