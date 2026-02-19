package service

import (
	"context"

	v1 "github.com/ZTH7/RagoDesk/apps/server/api/auth/v1"
	biz "github.com/ZTH7/RagoDesk/apps/server/internal/auth/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConsoleAuthService handles console auth endpoints.
type ConsoleAuthService struct {
	v1.UnimplementedConsoleAuthServer

	uc *biz.AuthUsecase
}

// PlatformAuthService handles platform auth endpoints.
type PlatformAuthService struct {
	v1.UnimplementedPlatformAuthServer

	uc *biz.AuthUsecase
}

// NewConsoleAuthService creates a new ConsoleAuthService.
func NewConsoleAuthService(uc *biz.AuthUsecase) *ConsoleAuthService {
	return &ConsoleAuthService{uc: uc}
}

// NewPlatformAuthService creates a new PlatformAuthService.
func NewPlatformAuthService(uc *biz.AuthUsecase) *PlatformAuthService {
	return &PlatformAuthService{uc: uc}
}

func (s *ConsoleAuthService) Login(ctx context.Context, req *v1.ConsoleLoginRequest) (*v1.AuthResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	session, err := s.uc.ConsoleLogin(ctx, req.GetAccount(), req.GetPassword(), req.GetTenantId())
	if err != nil {
		return nil, err
	}
	return toAuthResponse(session), nil
}

func (s *ConsoleAuthService) Register(ctx context.Context, req *v1.ConsoleRegisterRequest) (*v1.AuthResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	session, err := s.uc.ConsoleRegister(ctx, biz.ConsoleRegisterInput{
		TenantName: req.GetTenantName(),
		TenantType: req.GetTenantType(),
		AdminName:  req.GetAdminName(),
		Email:      req.GetEmail(),
		Phone:      req.GetPhone(),
		Password:   req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}
	return toAuthResponse(session), nil
}

func (s *PlatformAuthService) Login(ctx context.Context, req *v1.PlatformLoginRequest) (*v1.AuthResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	session, err := s.uc.PlatformLogin(ctx, req.GetAccount(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return toAuthResponse(session), nil
}

func toAuthResponse(session biz.AuthSession) *v1.AuthResponse {
	if session.Token == "" {
		return nil
	}
	return &v1.AuthResponse{
		Token:     session.Token,
		ExpiresAt: timestamppb.New(session.ExpiresAt),
		Profile: &v1.AuthProfile{
			SubjectId: session.Profile.SubjectID,
			TenantId:  session.Profile.TenantID,
			Account:   session.Profile.Account,
			Name:      session.Profile.Name,
			Roles:     session.Profile.Roles,
		},
	}
}

// ProviderSet is auth service providers.
var ProviderSet = wire.NewSet(NewConsoleAuthService, NewPlatformAuthService)
