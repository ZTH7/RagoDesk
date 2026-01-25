package service

import (
	"context"
	"time"

	v1 "github.com/ZTH7/RagoDesk/apps/server/api/bot/v1"
	botbiz "github.com/ZTH7/RagoDesk/apps/server/internal/bot/biz"
	iambiz "github.com/ZTH7/RagoDesk/apps/server/internal/iam/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BotService handles bot service layer.
type BotService struct {
	v1.UnimplementedConsoleBotServer

	uc    *botbiz.BotUsecase
	iamUC *iambiz.IAMUsecase
}

// NewBotService creates a new BotService.
func NewBotService(uc *botbiz.BotUsecase, iamUC *iambiz.IAMUsecase) *BotService {
	return &BotService{uc: uc, iamUC: iamUC}
}

func (s *BotService) CreateBot(ctx context.Context, req *v1.CreateBotRequest) (*v1.BotResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if s.iamUC != nil {
		if err := s.iamUC.RequirePermission(ctx, botbiz.PermissionBotWrite); err != nil {
			return nil, err
		}
	}
	created, err := s.uc.CreateBot(ctx, botbiz.Bot{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Status:      req.GetStatus(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.BotResponse{Bot: toBot(created)}, nil
}

func (s *BotService) GetBot(ctx context.Context, req *v1.GetBotRequest) (*v1.BotResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if s.iamUC != nil {
		if err := s.iamUC.RequirePermission(ctx, botbiz.PermissionBotRead); err != nil {
			return nil, err
		}
	}
	bot, err := s.uc.GetBot(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.BotResponse{Bot: toBot(bot)}, nil
}

func (s *BotService) ListBots(ctx context.Context, req *v1.ListBotsRequest) (*v1.ListBotsResponse, error) {
	if s.iamUC != nil {
		if err := s.iamUC.RequirePermission(ctx, botbiz.PermissionBotRead); err != nil {
			return nil, err
		}
	}
	bots, err := s.uc.ListBots(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.Bot, 0, len(bots))
	for _, bot := range bots {
		items = append(items, toBot(bot))
	}
	return &v1.ListBotsResponse{Items: items}, nil
}

func (s *BotService) UpdateBot(ctx context.Context, req *v1.UpdateBotRequest) (*v1.BotResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if s.iamUC != nil {
		if err := s.iamUC.RequirePermission(ctx, botbiz.PermissionBotWrite); err != nil {
			return nil, err
		}
	}
	updated, err := s.uc.UpdateBot(ctx, botbiz.Bot{
		ID:          req.GetId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Status:      req.GetStatus(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.BotResponse{Bot: toBot(updated)}, nil
}

func (s *BotService) DeleteBot(ctx context.Context, req *v1.DeleteBotRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if s.iamUC != nil {
		if err := s.iamUC.RequirePermission(ctx, botbiz.PermissionBotDelete); err != nil {
			return nil, err
		}
	}
	if err := s.uc.DeleteBot(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func toBot(bot botbiz.Bot) *v1.Bot {
	if bot.ID == "" && bot.Name == "" {
		return nil
	}
	return &v1.Bot{
		Id:          bot.ID,
		TenantId:    bot.TenantID,
		Name:        bot.Name,
		Description: bot.Description,
		Status:      bot.Status,
		CreatedAt:   timeOrNil(bot.CreatedAt),
		UpdatedAt:   timeOrNil(bot.UpdatedAt),
	}
}

func timeOrNil(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value)
}

// ProviderSet is bot service providers.
var ProviderSet = wire.NewSet(NewBotService)
