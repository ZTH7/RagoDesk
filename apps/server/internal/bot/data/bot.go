package data

import (
	"context"
	"database/sql"
	stderrors "errors"
	"strings"
	"time"

	internaldata "github.com/ZTH7/RagoDesk/apps/server/internal/data"
	biz "github.com/ZTH7/RagoDesk/apps/server/internal/bot/biz"
	"github.com/ZTH7/RagoDesk/apps/server/internal/kit/tenant"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/google/wire"
)

type botRepo struct {
	log *log.Helper
	db  *sql.DB
}

// NewBotRepo creates a new bot repo.
func NewBotRepo(data *internaldata.Data, logger log.Logger) biz.BotRepo {
	return &botRepo{log: log.NewHelper(logger), db: data.DB}
}

func (r *botRepo) CreateBot(ctx context.Context, bot biz.Bot) (biz.Bot, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.Bot{}, err
	}
	if bot.ID == "" {
		bot.ID = uuid.NewString()
	}
	if bot.Status == "" {
		bot.Status = "active"
	}
	bot.TenantID = tenantID
	now := time.Now()
	if bot.CreatedAt.IsZero() {
		bot.CreatedAt = now
	}
	bot.UpdatedAt = now
	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO bot (id, tenant_id, name, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		bot.ID,
		bot.TenantID,
		bot.Name,
		bot.Description,
		bot.Status,
		bot.CreatedAt,
		bot.UpdatedAt,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return biz.Bot{}, kerrors.Conflict("BOT_DUPLICATE", "bot already exists")
		}
		return biz.Bot{}, err
	}
	return bot, nil
}

func (r *botRepo) GetBot(ctx context.Context, id string) (biz.Bot, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.Bot{}, err
	}
	var bot biz.Bot
	err = r.db.QueryRowContext(
		ctx,
		"SELECT id, tenant_id, name, description, status, created_at, updated_at FROM bot WHERE tenant_id = ? AND id = ?",
		tenantID,
		id,
	).Scan(&bot.ID, &bot.TenantID, &bot.Name, &bot.Description, &bot.Status, &bot.CreatedAt, &bot.UpdatedAt)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return biz.Bot{}, kerrors.NotFound("BOT_NOT_FOUND", "bot not found")
		}
		return biz.Bot{}, err
	}
	return bot, nil
}

func (r *botRepo) ListBots(ctx context.Context, limit int, offset int) ([]biz.Bot, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, tenant_id, name, description, status, created_at, updated_at FROM bot WHERE tenant_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		tenantID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Bot, 0)
	for rows.Next() {
		var bot biz.Bot
		if err := rows.Scan(&bot.ID, &bot.TenantID, &bot.Name, &bot.Description, &bot.Status, &bot.CreatedAt, &bot.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, bot)
	}
	return items, rows.Err()
}

func (r *botRepo) UpdateBot(ctx context.Context, bot biz.Bot) (biz.Bot, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.Bot{}, err
	}
	current, err := r.GetBot(ctx, bot.ID)
	if err != nil {
		return biz.Bot{}, err
	}
	if strings.TrimSpace(bot.Name) == "" {
		bot.Name = current.Name
	}
	if strings.TrimSpace(bot.Description) == "" {
		bot.Description = current.Description
	}
	if strings.TrimSpace(bot.Status) == "" {
		bot.Status = current.Status
	}
	bot.TenantID = tenantID
	bot.CreatedAt = current.CreatedAt
	bot.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(
		ctx,
		"UPDATE bot SET name = ?, description = ?, status = ?, updated_at = ? WHERE tenant_id = ? AND id = ?",
		bot.Name,
		bot.Description,
		bot.Status,
		bot.UpdatedAt,
		tenantID,
		bot.ID,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return biz.Bot{}, kerrors.Conflict("BOT_DUPLICATE", "bot already exists")
		}
		return biz.Bot{}, err
	}
	return bot, nil
}

func (r *botRepo) DeleteBot(ctx context.Context, id string) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, "DELETE FROM bot WHERE tenant_id = ? AND id = ?", tenantID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return kerrors.NotFound("BOT_NOT_FOUND", "bot not found")
	}
	return nil
}

// ProviderSet is bot data providers.
var ProviderSet = wire.NewSet(NewBotRepo)
