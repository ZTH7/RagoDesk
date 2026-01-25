package biz

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// Bot represents a tenant bot.
type Bot struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permission codes for bot management.
const (
	PermissionBotRead   = "tenant.bot.read"
	PermissionBotWrite  = "tenant.bot.write"
	PermissionBotDelete = "tenant.bot.delete"
)

// BotRepo defines bot persistence operations.
type BotRepo interface {
	CreateBot(ctx context.Context, bot Bot) (Bot, error)
	GetBot(ctx context.Context, id string) (Bot, error)
	ListBots(ctx context.Context, limit int, offset int) ([]Bot, error)
	UpdateBot(ctx context.Context, bot Bot) (Bot, error)
	DeleteBot(ctx context.Context, id string) error
}

// BotUsecase handles bot domain logic.
type BotUsecase struct {
	repo BotRepo
	log  *log.Helper
}

// NewBotUsecase creates a new BotUsecase.
func NewBotUsecase(repo BotRepo, logger log.Logger) *BotUsecase {
	return &BotUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *BotUsecase) CreateBot(ctx context.Context, bot Bot) (Bot, error) {
	bot.Name = strings.TrimSpace(bot.Name)
	if bot.Name == "" {
		return Bot{}, errors.BadRequest("BOT_NAME_REQUIRED", "bot name required")
	}
	bot.Description = strings.TrimSpace(bot.Description)
	if bot.Status == "" {
		bot.Status = "active"
	} else if !isValidStatus(bot.Status) {
		return Bot{}, errors.BadRequest("BOT_STATUS_INVALID", "invalid bot status")
	}
	return uc.repo.CreateBot(ctx, bot)
}

func (uc *BotUsecase) GetBot(ctx context.Context, id string) (Bot, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Bot{}, errors.BadRequest("BOT_ID_REQUIRED", "bot id required")
	}
	return uc.repo.GetBot(ctx, id)
}

func (uc *BotUsecase) ListBots(ctx context.Context, limit int, offset int) ([]Bot, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return uc.repo.ListBots(ctx, limit, offset)
}

func (uc *BotUsecase) UpdateBot(ctx context.Context, bot Bot) (Bot, error) {
	bot.ID = strings.TrimSpace(bot.ID)
	if bot.ID == "" {
		return Bot{}, errors.BadRequest("BOT_ID_REQUIRED", "bot id required")
	}
	bot.Name = strings.TrimSpace(bot.Name)
	bot.Description = strings.TrimSpace(bot.Description)
	if bot.Status != "" && !isValidStatus(bot.Status) {
		return Bot{}, errors.BadRequest("BOT_STATUS_INVALID", "invalid bot status")
	}
	return uc.repo.UpdateBot(ctx, bot)
}

func (uc *BotUsecase) DeleteBot(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.BadRequest("BOT_ID_REQUIRED", "bot id required")
	}
	return uc.repo.DeleteBot(ctx, id)
}

func isValidStatus(status string) bool {
	return status == "active" || status == "disabled"
}

// ProviderSet is bot biz providers.
var ProviderSet = wire.NewSet(NewBotUsecase)
