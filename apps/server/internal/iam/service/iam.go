package service

import (
	"context"
	"time"

	v1 "github.com/ZTH7/RAGDesk/apps/server/api/iam/v1"
	"github.com/ZTH7/RAGDesk/apps/server/internal/auth"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/iam/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IAMService handles iam service layer (placeholder)
type IAMService struct {
	v1.UnimplementedAdminIAMServer
	uc  *biz.IAMUsecase
	log *log.Helper
}

// NewIAMService creates a new IAMService
func NewIAMService(uc *biz.IAMUsecase, logger log.Logger) *IAMService {
	return &IAMService{uc: uc, log: log.NewHelper(logger)}
}

func (s *IAMService) CreateTenant(ctx context.Context, req *v1.CreateTenantRequest) (*v1.TenantResponse, error) {
	if err := s.uc.RequirePermission(ctx, biz.PermissionTenantCreate); err != nil {
		return nil, err
	}
	tenantModel := biz.Tenant{
		Name:   req.GetName(),
		Plan:   req.GetPlan(),
		Status: req.GetStatus(),
	}
	created, err := s.uc.CreateTenant(ctx, tenantModel)
	if err != nil {
		return nil, err
	}
	return &v1.TenantResponse{Tenant: toTenant(created)}, nil
}

func (s *IAMService) GetTenant(ctx context.Context, req *v1.GetTenantRequest) (*v1.TenantResponse, error) {
	if err := s.uc.RequirePermission(ctx, biz.PermissionTenantRead); err != nil {
		return nil, err
	}
	tenantModel, err := s.uc.GetTenant(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.TenantResponse{Tenant: toTenant(tenantModel)}, nil
}

func (s *IAMService) ListTenants(ctx context.Context, req *v1.ListTenantsRequest) (*v1.ListTenantsResponse, error) {
	if err := s.uc.RequirePermission(ctx, biz.PermissionTenantRead); err != nil {
		return nil, err
	}
	tenants, err := s.uc.ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	resp := &v1.ListTenantsResponse{Items: make([]*v1.Tenant, 0, len(tenants))}
	for _, item := range tenants {
		resp.Items = append(resp.Items, toTenant(item))
	}
	return resp, nil
}

func (s *IAMService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.UserResponse, error) {
	ctx, err := ensureTenantScope(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}
	if err := s.uc.RequirePermission(ctx, biz.PermissionUserWrite); err != nil {
		return nil, err
	}
	userModel := biz.User{
		Email:  req.GetEmail(),
		Phone:  req.GetPhone(),
		Name:   req.GetName(),
		Status: req.GetStatus(),
	}
	created, err := s.uc.CreateUser(ctx, userModel)
	if err != nil {
		return nil, err
	}
	return &v1.UserResponse{User: toUser(created)}, nil
}

func (s *IAMService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	ctx, err := ensureTenantScope(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}
	if err := s.uc.RequirePermission(ctx, biz.PermissionUserRead); err != nil {
		return nil, err
	}
	users, err := s.uc.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	resp := &v1.ListUsersResponse{Items: make([]*v1.User, 0, len(users))}
	for _, item := range users {
		resp.Items = append(resp.Items, toUser(item))
	}
	return resp, nil
}

func (s *IAMService) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.RoleResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.uc.RequirePermission(ctx, biz.PermissionRoleWrite); err != nil {
		return nil, err
	}
	roleModel := biz.Role{
		Name: req.GetName(),
	}
	created, err := s.uc.CreateRole(ctx, roleModel)
	if err != nil {
		return nil, err
	}
	return &v1.RoleResponse{Role: toRole(created)}, nil
}

func (s *IAMService) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.uc.RequirePermission(ctx, biz.PermissionRoleRead); err != nil {
		return nil, err
	}
	roles, err := s.uc.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	resp := &v1.ListRolesResponse{Items: make([]*v1.Role, 0, len(roles))}
	for _, item := range roles {
		resp.Items = append(resp.Items, toRole(item))
	}
	return resp, nil
}

func (s *IAMService) AssignRole(ctx context.Context, req *v1.AssignRoleRequest) (*emptypb.Empty, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.uc.RequirePermission(ctx, biz.PermissionRoleAssign); err != nil {
		return nil, err
	}
	if err := s.uc.AssignRole(ctx, req.GetUserId(), req.GetRoleId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *IAMService) CreatePermission(ctx context.Context, req *v1.CreatePermissionRequest) (*v1.PermissionResponse, error) {
	if err := s.uc.RequirePermission(ctx, biz.PermissionPermissionWrite); err != nil {
		return nil, err
	}
	permission := biz.Permission{
		Code:        req.GetCode(),
		Description: req.GetDescription(),
	}
	created, err := s.uc.CreatePermission(ctx, permission)
	if err != nil {
		return nil, err
	}
	return &v1.PermissionResponse{Permission: toPermission(created)}, nil
}

func (s *IAMService) ListPermissions(ctx context.Context, req *v1.ListPermissionsRequest) (*v1.ListPermissionsResponse, error) {
	if err := s.uc.RequirePermission(ctx, biz.PermissionPermissionRead); err != nil {
		return nil, err
	}
	permissions, err := s.uc.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	resp := &v1.ListPermissionsResponse{Items: make([]*v1.Permission, 0, len(permissions))}
	for _, item := range permissions {
		resp.Items = append(resp.Items, toPermission(item))
	}
	return resp, nil
}

func (s *IAMService) AssignRolePermissions(ctx context.Context, req *v1.AssignRolePermissionsRequest) (*emptypb.Empty, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.uc.RequirePermission(ctx, biz.PermissionRolePermAssign); err != nil {
		return nil, err
	}
	if err := s.uc.AssignRolePermissions(ctx, req.GetRoleId(), req.GetPermissionCodes()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *IAMService) ListRolePermissions(ctx context.Context, req *v1.ListRolePermissionsRequest) (*v1.ListPermissionsResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.uc.RequirePermission(ctx, biz.PermissionRoleRead); err != nil {
		return nil, err
	}
	permissions, err := s.uc.ListRolePermissions(ctx, req.GetRoleId())
	if err != nil {
		return nil, err
	}
	resp := &v1.ListPermissionsResponse{Items: make([]*v1.Permission, 0, len(permissions))}
	for _, item := range permissions {
		resp.Items = append(resp.Items, toPermission(item))
	}
	return resp, nil
}

func ensureTenantScope(ctx context.Context, tenantID string) (context.Context, error) {
	if tenantID == "" {
		return ctx, nil
	}
	claims, _ := auth.ClaimsFromContext(ctx)
	if claims != nil && claims.TenantID != "" && claims.TenantID != tenantID {
		return nil, errors.Forbidden("TENANT_MISMATCH", "tenant mismatch")
	}
	return tenant.WithTenantID(ctx, tenantID), nil
}

func requireTenantContext(ctx context.Context) error {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return errors.Forbidden("TENANT_MISSING", "tenant missing")
	}
	return nil
}

func toTimestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value)
}

func toTenant(value biz.Tenant) *v1.Tenant {
	if value.ID == "" && value.Name == "" && value.Plan == "" && value.Status == "" {
		return nil
	}
	return &v1.Tenant{
		Id:        value.ID,
		Name:      value.Name,
		Plan:      value.Plan,
		Status:    value.Status,
		CreatedAt: toTimestamp(value.CreatedAt),
	}
}

func toUser(value biz.User) *v1.User {
	if value.ID == "" && value.Email == "" && value.Phone == "" && value.Name == "" {
		return nil
	}
	return &v1.User{
		Id:        value.ID,
		TenantId:  value.TenantID,
		Email:     value.Email,
		Phone:     value.Phone,
		Name:      value.Name,
		Status:    value.Status,
		CreatedAt: toTimestamp(value.CreatedAt),
	}
}

func toRole(value biz.Role) *v1.Role {
	if value.ID == "" && value.Name == "" {
		return nil
	}
	return &v1.Role{
		Id:       value.ID,
		TenantId: value.TenantID,
		Name:     value.Name,
	}
}

func toPermission(value biz.Permission) *v1.Permission {
	if value.ID == "" && value.Code == "" && value.Description == "" {
		return nil
	}
	return &v1.Permission{
		Id:          value.ID,
		Code:        value.Code,
		Description: value.Description,
	}
}

// ProviderSet is iam service providers.
var ProviderSet = wire.NewSet(NewIAMService)
