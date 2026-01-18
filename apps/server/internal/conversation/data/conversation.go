package data

import (
	"context"
	"database/sql"
	"strings"
	"time"

	internaldata "github.com/ZTH7/RAGDesk/apps/server/internal/data"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"
)

type conversationRepo struct {
	db *sql.DB
}

// NewConversationRepo creates a new conversation repo.
func NewConversationRepo(data *internaldata.Data) biz.ConversationRepo {
	return &conversationRepo{db: data.DB}
}

func (r *conversationRepo) CreateSession(ctx context.Context, session biz.Session) (biz.Session, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.Session{}, err
	}
	session.TenantID = tenantID
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = session.CreatedAt
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO chat_session
			(id, tenant_id, bot_id, status, close_reason, user_external_id, metadata, created_at, updated_at, closed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID,
		session.TenantID,
		session.BotID,
		session.Status,
		session.CloseReason,
		session.UserExternal,
		session.Metadata,
		session.CreatedAt,
		session.UpdatedAt,
		nullTime(session.ClosedAt),
	)
	if err != nil {
		return biz.Session{}, err
	}
	return session, nil
}

func (r *conversationRepo) GetSession(ctx context.Context, sessionID string) (biz.Session, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.Session{}, err
	}
	var s biz.Session
	err = r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, bot_id, status, close_reason, user_external_id, metadata, created_at, updated_at, closed_at
		FROM chat_session WHERE tenant_id = ? AND id = ?`,
		tenantID,
		sessionID,
	).Scan(
		&s.ID,
		&s.TenantID,
		&s.BotID,
		&s.Status,
		&s.CloseReason,
		&s.UserExternal,
		&s.Metadata,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.ClosedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return biz.Session{}, errors.NotFound("SESSION_NOT_FOUND", "session not found")
		}
		return biz.Session{}, err
	}
	return s, nil
}

func (r *conversationRepo) CloseSession(ctx context.Context, sessionID string, closeReason string, closedAt time.Time) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	if closedAt.IsZero() {
		closedAt = time.Now()
	}
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE chat_session
		 SET status = ?, close_reason = ?, updated_at = ?, closed_at = ?
		 WHERE tenant_id = ? AND id = ?`,
		biz.SessionStatusClosed,
		closeReason,
		closedAt,
		closedAt,
		tenantID,
		sessionID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return errors.NotFound("SESSION_NOT_FOUND", "session not found")
	}
	return err
}

func (r *conversationRepo) PurgeExpired(ctx context.Context, cutoff time.Time) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	if cutoff.IsZero() {
		return nil
	}
	_, _ = r.db.ExecContext(
		ctx,
		`DELETE FROM message_feedback
		 WHERE tenant_id = ? AND session_id IN (
			SELECT id FROM chat_session WHERE tenant_id = ? AND created_at < ?
		 )`,
		tenantID,
		tenantID,
		cutoff,
	)
	_, _ = r.db.ExecContext(
		ctx,
		`DELETE FROM session_event
		 WHERE tenant_id = ? AND session_id IN (
			SELECT id FROM chat_session WHERE tenant_id = ? AND created_at < ?
		 )`,
		tenantID,
		tenantID,
		cutoff,
	)
	_, _ = r.db.ExecContext(
		ctx,
		`DELETE FROM chat_message
		 WHERE tenant_id = ? AND session_id IN (
			SELECT id FROM chat_session WHERE tenant_id = ? AND created_at < ?
		 )`,
		tenantID,
		tenantID,
		cutoff,
	)
	_, err = r.db.ExecContext(
		ctx,
		`DELETE FROM chat_session WHERE tenant_id = ? AND created_at < ?`,
		tenantID,
		cutoff,
	)
	return err
}

func (r *conversationRepo) ListSessions(ctx context.Context, botID string, limit int, offset int) ([]biz.Session, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	query := `SELECT id, tenant_id, bot_id, status, close_reason, user_external_id, metadata, created_at, updated_at, closed_at
		FROM chat_session WHERE tenant_id = ?`
	args := []any{tenantID}
	if strings.TrimSpace(botID) != "" {
		query += " AND bot_id = ?"
		args = append(args, strings.TrimSpace(botID))
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := make([]biz.Session, 0)
	for rows.Next() {
		var s biz.Session
		if err := rows.Scan(
			&s.ID,
			&s.TenantID,
			&s.BotID,
			&s.Status,
			&s.CloseReason,
			&s.UserExternal,
			&s.Metadata,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.ClosedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *conversationRepo) CreateMessages(ctx context.Context, messages []biz.Message) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	if len(messages) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, msg := range messages {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO chat_message
				(id, tenant_id, session_id, role, content, confidence, references_json, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			msg.ID,
			tenantID,
			msg.SessionID,
			msg.Role,
			msg.Content,
			msg.Confidence,
			msg.References,
			msg.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *conversationRepo) ListMessages(ctx context.Context, sessionID string, limit int, offset int) ([]biz.Message, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, session_id, role, content, confidence, references_json, created_at
		FROM chat_message WHERE tenant_id = ? AND session_id = ?
		ORDER BY created_at ASC LIMIT ? OFFSET ?`,
		tenantID,
		sessionID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Message, 0)
	for rows.Next() {
		var m biz.Message
		if err := rows.Scan(
			&m.ID,
			&m.TenantID,
			&m.SessionID,
			&m.Role,
			&m.Content,
			&m.Confidence,
			&m.References,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *conversationRepo) CreateEvent(ctx context.Context, event biz.SessionEvent) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO session_event (id, tenant_id, session_id, event_type, event_detail, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		event.ID,
		tenantID,
		event.SessionID,
		event.EventType,
		event.Detail,
		event.CreatedAt,
	)
	return err
}

func (r *conversationRepo) CreateFeedback(ctx context.Context, feedback biz.MessageFeedback) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	if feedback.CreatedAt.IsZero() {
		feedback.CreatedAt = time.Now()
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO message_feedback
			(id, tenant_id, session_id, message_id, rating, comment, correction, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		feedback.ID,
		tenantID,
		feedback.SessionID,
		feedback.MessageID,
		feedback.Rating,
		feedback.Comment,
		feedback.Correction,
		feedback.CreatedAt,
	)
	return err
}

// ProviderSet is conversation data providers.
var ProviderSet = wire.NewSet(NewConversationRepo)

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
