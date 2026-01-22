package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"

	_ "github.com/go-sql-driver/mysql"
)

// Data .
type Data struct {
	DB *sql.DB
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	if c == nil || c.Database == nil || c.Database.Driver == "" || c.Database.Source == "" {
		return nil, nil, errors.InternalServer("DB_CONFIG_MISSING", "database config missing")
	}
	db, err := sql.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	db.SetConnMaxLifetime(1 * time.Hour)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(50)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	schemaCtx, schemaCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer schemaCancel()
	if err := ensureSchemaAndSeed(schemaCtx, db); err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	cleanup := func() {
		log.Info("closing the data resources")
		if err := db.Close(); err != nil {
			log.Errorf("close database error: %v", err)
		}
	}
	return &Data{DB: db}, cleanup, nil
}

func ensureSchemaAndSeed(ctx context.Context, db *sql.DB) error {
	if err := ensureIAMSchema(ctx, db); err != nil {
		return err
	}
	if err := ensureAPIMgmtSchema(ctx, db); err != nil {
		return err
	}
	if err := ensureKnowledgeSchema(ctx, db); err != nil {
		return err
	}
	if err := ensureConversationSchema(ctx, db); err != nil {
		return err
	}
	if err := seedIAMPermissions(ctx, db); err != nil {
		return err
	}
	return nil
}

func ensureAPIMgmtSchema(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS api_key (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			bot_id VARCHAR(36) NOT NULL,
			name VARCHAR(128) NOT NULL,
			key_hash VARCHAR(64) NOT NULL,
			scopes TEXT NULL,
			api_versions TEXT NULL,
			prev_key_hash VARCHAR(64) NULL,
			prev_expires_at DATETIME NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			quota_daily INT NOT NULL DEFAULT 0,
			qps_limit INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			last_used_at DATETIME NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_api_key_hash (key_hash),
			KEY idx_api_key_tenant (tenant_id),
			KEY idx_api_key_bot (tenant_id, bot_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS api_usage_log (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			bot_id VARCHAR(36) NOT NULL,
			api_key_id VARCHAR(36) NOT NULL,
			path VARCHAR(255) NOT NULL,
			api_version VARCHAR(16) NULL,
			model VARCHAR(128) NULL,
			status_code INT NOT NULL,
			latency_ms INT NOT NULL,
			prompt_tokens INT NOT NULL DEFAULT 0,
			completion_tokens INT NOT NULL DEFAULT 0,
			total_tokens INT NOT NULL DEFAULT 0,
			client_ip VARCHAR(64) NULL,
			user_agent VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			KEY idx_api_usage_tenant (tenant_id),
			KEY idx_api_usage_key (api_key_id),
			KEY idx_api_usage_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if err := ensureColumn(ctx, db, "api_key", "bot_id", "VARCHAR(36) NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "key_hash", "VARCHAR(64) NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "scopes", "TEXT NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "api_versions", "TEXT NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "prev_key_hash", "VARCHAR(64) NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "prev_expires_at", "DATETIME NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "status", "VARCHAR(32) NOT NULL DEFAULT 'active'"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "quota_daily", "INT NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "qps_limit", "INT NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_key", "last_used_at", "DATETIME NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "tenant_id", "VARCHAR(36) NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "bot_id", "VARCHAR(36) NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "api_version", "VARCHAR(16) NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "model", "VARCHAR(128) NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "prompt_tokens", "INT NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "completion_tokens", "INT NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "total_tokens", "INT NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "client_ip", "VARCHAR(64) NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_usage_log", "user_agent", "VARCHAR(255) NULL"); err != nil {
		return err
	}
	return nil
}

func ensureIAMSchema(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS tenant (
			id VARCHAR(36) NOT NULL,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(32) NOT NULL DEFAULT 'enterprise',
			plan VARCHAR(32) NOT NULL DEFAULT 'free',
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			KEY idx_tenant_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS ` + "`user`" + ` (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			email VARCHAR(255) NULL,
			phone VARCHAR(32) NULL,
			name VARCHAR(255) NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_user_tenant_email (tenant_id, email),
			KEY idx_user_tenant (tenant_id),
			KEY idx_user_created_at (tenant_id, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS ` + "`role`" + ` (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			name VARCHAR(128) NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_role_tenant_name (tenant_id, name),
			KEY idx_role_tenant (tenant_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS permission (
			id VARCHAR(36) NOT NULL,
			code VARCHAR(128) NOT NULL,
			description VARCHAR(255) NOT NULL,
			scope VARCHAR(32) NOT NULL DEFAULT 'platform',
			PRIMARY KEY (id),
			UNIQUE KEY uniq_permission_code (code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS user_role (
			user_id VARCHAR(36) NOT NULL,
			role_id VARCHAR(36) NOT NULL,
			UNIQUE KEY uniq_user_role (user_id, role_id),
			KEY idx_user_role_user (user_id),
			KEY idx_user_role_role (role_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS role_permission (
			role_id VARCHAR(36) NOT NULL,
			permission_id VARCHAR(36) NOT NULL,
			UNIQUE KEY uniq_role_permission (role_id, permission_id),
			KEY idx_role_permission_role (role_id),
			KEY idx_role_permission_permission (permission_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS platform_admin (
			id VARCHAR(36) NOT NULL,
			email VARCHAR(255) NULL,
			phone VARCHAR(32) NULL,
			name VARCHAR(255) NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			password_hash VARCHAR(255) NOT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_platform_admin_email (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS platform_role (
			id VARCHAR(36) NOT NULL,
			name VARCHAR(128) NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_platform_role_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS platform_admin_role (
			admin_id VARCHAR(36) NOT NULL,
			role_id VARCHAR(36) NOT NULL,
			UNIQUE KEY uniq_platform_admin_role (admin_id, role_id),
			KEY idx_platform_admin_role_admin (admin_id),
			KEY idx_platform_admin_role_role (role_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS platform_role_permission (
			role_id VARCHAR(36) NOT NULL,
			permission_id VARCHAR(36) NOT NULL,
			UNIQUE KEY uniq_platform_role_permission (role_id, permission_id),
			KEY idx_platform_role_permission_role (role_id),
			KEY idx_platform_role_permission_permission (permission_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	if err := ensureColumn(ctx, db, "tenant", "type", "VARCHAR(32) NOT NULL DEFAULT 'enterprise'"); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, "UPDATE tenant SET `type` = 'enterprise' WHERE `type` IS NULL OR `type` = ''"); err != nil {
		return err
	}

	if err := ensureColumn(ctx, db, "permission", "scope", "VARCHAR(32) NOT NULL DEFAULT 'platform'"); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, "UPDATE permission SET scope = 'platform' WHERE scope IS NULL OR scope = ''"); err != nil {
		return err
	}
	return nil
}

func ensureKnowledgeSchema(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS knowledge_base (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_kb_tenant_name (tenant_id, name),
			KEY idx_kb_tenant (tenant_id),
			KEY idx_kb_created_at (tenant_id, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS document (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			kb_id VARCHAR(36) NOT NULL,
			title VARCHAR(255) NOT NULL,
			source_type VARCHAR(32) NOT NULL,
			status VARCHAR(32) NOT NULL,
			current_version INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			KEY idx_document_tenant_kb (tenant_id, kb_id),
			KEY idx_document_created_at (tenant_id, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS document_version (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			document_id VARCHAR(36) NOT NULL,
			version INT NOT NULL,
			raw_uri VARCHAR(1024) NULL,
			index_config_hash VARCHAR(64) NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL,
			error_message TEXT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_doc_version (document_id, version),
			KEY idx_doc_version_tenant_doc (tenant_id, document_id),
			KEY idx_doc_version_created_at (tenant_id, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS doc_chunk (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			kb_id VARCHAR(36) NOT NULL,
			document_id VARCHAR(36) NOT NULL,
			document_version_id VARCHAR(36) NOT NULL,
			chunk_index INT NOT NULL,
			content LONGTEXT NOT NULL,
			token_count INT NOT NULL,
			content_hash VARCHAR(64) NOT NULL,
			language VARCHAR(32) NOT NULL DEFAULT '',
			section VARCHAR(255) NULL,
			page_no INT NULL,
			source_uri VARCHAR(1024) NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_chunk_version_index (document_version_id, chunk_index),
			KEY idx_chunk_tenant_kb (tenant_id, kb_id),
			KEY idx_chunk_version (tenant_id, document_version_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS embedding (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			chunk_id VARCHAR(36) NOT NULL,
			model VARCHAR(128) NOT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_embedding_chunk_model (chunk_id, model),
			KEY idx_embedding_tenant_chunk (tenant_id, chunk_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS bot_kb (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			bot_id VARCHAR(36) NOT NULL,
			kb_id VARCHAR(36) NOT NULL,
			weight DOUBLE NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_bot_kb (tenant_id, bot_id, kb_id),
			KEY idx_bot_kb_tenant_bot (tenant_id, bot_id),
			KEY idx_bot_kb_tenant_kb (tenant_id, kb_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if err := ensureDropColumn(ctx, db, "document_version", "raw_text"); err != nil {
		return err
	}
	if err := ensureDropColumn(ctx, db, "bot_kb", "priority"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "document_version", "raw_uri", "VARCHAR(1024) NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "document_version", "index_config_hash", "VARCHAR(64) NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "doc_chunk", "section", "VARCHAR(255) NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "doc_chunk", "page_no", "INT NULL"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "doc_chunk", "source_uri", "VARCHAR(1024) NULL"); err != nil {
		return err
	}
	return nil
}

func ensureConversationSchema(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS chat_session (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			bot_id VARCHAR(36) NOT NULL,
			status VARCHAR(32) NOT NULL,
			close_reason VARCHAR(64) NULL,
			user_external_id VARCHAR(128) NULL,
			metadata TEXT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			closed_at DATETIME NULL,
			PRIMARY KEY (id),
			KEY idx_chat_session_tenant (tenant_id),
			KEY idx_chat_session_tenant_bot (tenant_id, bot_id),
			KEY idx_chat_session_created_at (tenant_id, created_at),
			KEY idx_chat_session_status (tenant_id, status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS chat_message (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			session_id VARCHAR(36) NOT NULL,
			role VARCHAR(16) NOT NULL,
			content LONGTEXT NOT NULL,
			confidence DOUBLE NOT NULL DEFAULT 0,
			references_json TEXT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			KEY idx_chat_message_session (tenant_id, session_id, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS session_event (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			session_id VARCHAR(36) NOT NULL,
			event_type VARCHAR(32) NOT NULL,
			event_detail TEXT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			KEY idx_session_event_session (tenant_id, session_id, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS message_feedback (
			id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			session_id VARCHAR(36) NOT NULL,
			message_id VARCHAR(36) NOT NULL,
			rating INT NOT NULL,
			comment TEXT NULL,
			correction TEXT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			KEY idx_message_feedback_message (tenant_id, message_id),
			KEY idx_message_feedback_session (tenant_id, session_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func ensureColumn(ctx context.Context, db *sql.DB, table string, column string, definition string) error {
	var count int
	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		table,
		column,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	query := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", table, column, definition)
	_, err = db.ExecContext(ctx, query)
	return err
}

func ensureDropColumn(ctx context.Context, db *sql.DB, table string, column string) error {
	var count int
	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		table,
		column,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	query := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", table, column)
	_, err = db.ExecContext(ctx, query)
	return err
}

type permissionSeed struct {
	code        string
	description string
	scope       string
}

func seedIAMPermissions(ctx context.Context, db *sql.DB) error {
	seeds := []permissionSeed{
		{code: "platform.tenant.create", description: "Create tenant", scope: "platform"},
		{code: "platform.tenant.read", description: "Read tenant", scope: "platform"},
		{code: "platform.tenant.write", description: "Update tenant plan/status/quota", scope: "platform"},
		{code: "platform.admin.create", description: "Create platform admin", scope: "platform"},
		{code: "platform.admin.read", description: "Read platform admin", scope: "platform"},
		{code: "platform.role.write", description: "Create/update platform role", scope: "platform"},
		{code: "platform.role.read", description: "Read platform role", scope: "platform"},
		{code: "platform.role.assign", description: "Assign platform role to admin", scope: "platform"},
		{code: "platform.role.permission.assign", description: "Assign permissions to platform role", scope: "platform"},
		{code: "platform.permission.read", description: "Read permission catalog", scope: "platform"},
		{code: "platform.permission.write", description: "Create permission", scope: "platform"},
		{code: "platform.config.read", description: "Read platform configuration", scope: "platform"},
		{code: "platform.config.write", description: "Update platform configuration", scope: "platform"},
		{code: "tenant.user.read", description: "Read tenant users", scope: "tenant"},
		{code: "tenant.user.write", description: "Create/update tenant users", scope: "tenant"},
		{code: "tenant.role.read", description: "Read tenant roles", scope: "tenant"},
		{code: "tenant.role.write", description: "Create/update tenant roles", scope: "tenant"},
		{code: "tenant.role.assign", description: "Assign role to user", scope: "tenant"},
		{code: "tenant.role.permission.assign", description: "Assign permissions to role", scope: "tenant"},
		{code: "tenant.permission.read", description: "Read tenant permission catalog", scope: "tenant"},
		{code: "tenant.bot.read", description: "Read bots", scope: "tenant"},
		{code: "tenant.bot.write", description: "Create/update bots", scope: "tenant"},
		{code: "tenant.bot.delete", description: "Delete bots", scope: "tenant"},
		{code: "tenant.bot_kb.bind", description: "Bind bot to knowledge base", scope: "tenant"},
		{code: "tenant.bot_kb.unbind", description: "Unbind bot from knowledge base", scope: "tenant"},
		{code: "tenant.knowledge_base.read", description: "Read knowledge bases", scope: "tenant"},
		{code: "tenant.knowledge_base.write", description: "Create/update knowledge bases", scope: "tenant"},
		{code: "tenant.knowledge_base.delete", description: "Delete knowledge bases", scope: "tenant"},
		{code: "tenant.document.upload", description: "Upload documents", scope: "tenant"},
		{code: "tenant.document.read", description: "Read documents", scope: "tenant"},
		{code: "tenant.document.delete", description: "Delete documents", scope: "tenant"},
		{code: "tenant.document.reindex", description: "Reindex documents", scope: "tenant"},
		{code: "tenant.document.rollback", description: "Rollback document versions", scope: "tenant"},
		{code: "tenant.api_key.read", description: "Read API keys", scope: "tenant"},
		{code: "tenant.api_key.write", description: "Create/update API keys", scope: "tenant"},
		{code: "tenant.api_key.delete", description: "Delete API keys", scope: "tenant"},
		{code: "tenant.api_key.rotate", description: "Rotate API keys", scope: "tenant"},
		{code: "tenant.api_usage.read", description: "Read API usage logs", scope: "tenant"},
		{code: "tenant.analytics.read", description: "Read analytics dashboard", scope: "tenant"},
		{code: "tenant.chat_session.read", description: "Read chat sessions", scope: "tenant"},
		{code: "tenant.chat_message.read", description: "Read chat messages", scope: "tenant"},
	}

	for _, item := range seeds {
		if _, err := db.ExecContext(
			ctx,
			"INSERT IGNORE INTO permission (id, code, description, scope) VALUES (UUID(), ?, ?, ?)",
			item.code,
			item.description,
			item.scope,
		); err != nil {
			return err
		}
	}
	return nil
}
