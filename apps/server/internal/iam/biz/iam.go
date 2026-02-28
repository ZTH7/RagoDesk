package biz

import (
	"context"
	"strings"
	"time"

	jwt "github.com/ZTH7/RagoDesk/apps/server/internal/kit/jwt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// IAM domain model (placeholder)
type IAM struct {
	ID string
}

// Tenant represents a company tenant.
type Tenant struct {
	ID        string
	Name      string
	Type      string
	Plan      string
	Status    string
	CreatedAt time.Time
}

// User represents a tenant user.
type User struct {
	ID        string
	TenantID  string
	Email     string
	Phone     string
	Name      string
	Status    string
	CreatedAt time.Time
}

// Role represents a tenant role.
type Role struct {
	ID       string
	TenantID string
	Name     string
}

// Permission represents a permission code.
type Permission struct {
	ID          string
	Code        string
	Description string
	Scope       string
}

// UserRole ties a user to a role.
type UserRole struct {
	UserID string
	RoleID string
}

// PlatformAdmin represents a platform admin account.
type PlatformAdmin struct {
	ID           string
	Email        string
	Phone        string
	Name         string
	Status       string
	PasswordHash string
	CreatedAt    time.Time
}

// PlatformRole represents a platform role.
type PlatformRole struct {
	ID   string
	Name string
}

// Permission scopes.
const (
	PermissionScopePlatform = "platform"
	PermissionScopeTenant   = "tenant"
)

// Permission codes for IAM RBAC.
const (
	PermissionTenantCreate            = "platform.tenant.create"
	PermissionTenantRead              = "platform.tenant.read"
	PermissionPlatformAdminCreate     = "platform.admin.create"
	PermissionPlatformAdminRead       = "platform.admin.read"
	PermissionPlatformRoleWrite       = "platform.role.write"
	PermissionPlatformRoleRead        = "platform.role.read"
	PermissionPlatformRoleAssign      = "platform.role.assign"
	PermissionPlatformRolePermAssign  = "platform.role.permission.assign"
	PermissionPlatformPermissionRead  = "platform.permission.read"
	PermissionPlatformPermissionWrite = "platform.permission.write"
	PermissionUserWrite               = "tenant.user.write"
	PermissionUserRead                = "tenant.user.read"
	PermissionRoleWrite               = "tenant.role.write"
	PermissionRoleRead                = "tenant.role.read"
	PermissionRoleAssign              = "tenant.role.assign"
	PermissionTenantPermissionRead    = "tenant.permission.read"
	PermissionRolePermAssign          = "tenant.role.permission.assign"
)

// IAMRepo is a repository interface (placeholder)
type IAMRepo interface {
	Ping(context.Context) error
	CreateTenant(ctx context.Context, tenant Tenant) (Tenant, error)
	GetTenant(ctx context.Context, id string) (Tenant, error)
	ListTenants(ctx context.Context) ([]Tenant, error)

	CreatePlatformAdmin(ctx context.Context, admin PlatformAdmin) (PlatformAdmin, error)
	ListPlatformAdmins(ctx context.Context) ([]PlatformAdmin, error)
	GetPlatformAdmin(ctx context.Context, id string) (PlatformAdmin, error)
	CreatePlatformRole(ctx context.Context, role PlatformRole) (PlatformRole, error)
	ListPlatformRoles(ctx context.Context) ([]PlatformRole, error)
	GetPlatformRole(ctx context.Context, id string) (PlatformRole, error)
	AssignPlatformAdminRole(ctx context.Context, adminID string, roleID string) error
	ListPlatformAdminRoles(ctx context.Context, adminID string) ([]PlatformRole, error)
	RemovePlatformAdminRole(ctx context.Context, adminID string, roleID string) error
	ListPlatformAdminPermissions(ctx context.Context, adminID string) ([]Permission, error)
	AssignPlatformRolePermissions(ctx context.Context, roleID string, permissionCodes []string) error
	ListPlatformRolePermissions(ctx context.Context, roleID string) ([]Permission, error)

	CreateUser(ctx context.Context, user User) (User, error)
	ListUsers(ctx context.Context) ([]User, error)

	CreateRole(ctx context.Context, role Role) (Role, error)
	ListRoles(ctx context.Context) ([]Role, error)

	AssignRole(ctx context.Context, userID string, roleID string) error
	ListUserRoles(ctx context.Context, userID string) ([]Role, error)
	ListUserPermissions(ctx context.Context, userID string) ([]Permission, error)

	CreatePermission(ctx context.Context, permission Permission) (Permission, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	ListTenantPermissions(ctx context.Context) ([]Permission, error)
	AssignRolePermissions(ctx context.Context, roleID string, permissionCodes []string) error
	ListRolePermissions(ctx context.Context, roleID string) ([]Permission, error)
}

// IAMUsecase handles iam business logic (placeholder)
type IAMUsecase struct {
	repo IAMRepo
	log  *log.Helper
}

// NewIAMUsecase creates a new IAMUsecase
func NewIAMUsecase(repo IAMRepo, logger log.Logger) *IAMUsecase {
	return &IAMUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateTenant creates a tenant (platform admin).
func (uc *IAMUsecase) CreateTenant(ctx context.Context, tenant Tenant) (Tenant, error) {
	return uc.repo.CreateTenant(ctx, tenant)
}

// GetTenant returns tenant by id (platform admin).
func (uc *IAMUsecase) GetTenant(ctx context.Context, id string) (Tenant, error) {
	return uc.repo.GetTenant(ctx, id)
}

// ListTenants lists all tenants (platform admin).
func (uc *IAMUsecase) ListTenants(ctx context.Context) ([]Tenant, error) {
	return uc.repo.ListTenants(ctx)
}

// CreatePlatformAdmin creates a platform admin account.
func (uc *IAMUsecase) CreatePlatformAdmin(ctx context.Context, admin PlatformAdmin) (PlatformAdmin, error) {
	return uc.repo.CreatePlatformAdmin(ctx, admin)
}

// ListPlatformAdmins lists platform admins.
func (uc *IAMUsecase) ListPlatformAdmins(ctx context.Context) ([]PlatformAdmin, error) {
	return uc.repo.ListPlatformAdmins(ctx)
}

// GetPlatformAdmin returns platform admin by id.
func (uc *IAMUsecase) GetPlatformAdmin(ctx context.Context, id string) (PlatformAdmin, error) {
	return uc.repo.GetPlatformAdmin(ctx, id)
}

// CreatePlatformRole creates a platform role.
func (uc *IAMUsecase) CreatePlatformRole(ctx context.Context, role PlatformRole) (PlatformRole, error) {
	return uc.repo.CreatePlatformRole(ctx, role)
}

// ListPlatformRoles lists platform roles.
func (uc *IAMUsecase) ListPlatformRoles(ctx context.Context) ([]PlatformRole, error) {
	return uc.repo.ListPlatformRoles(ctx)
}

// GetPlatformRole returns platform role by id.
func (uc *IAMUsecase) GetPlatformRole(ctx context.Context, id string) (PlatformRole, error) {
	return uc.repo.GetPlatformRole(ctx, id)
}

// AssignPlatformAdminRole assigns a role to a platform admin.
func (uc *IAMUsecase) AssignPlatformAdminRole(ctx context.Context, adminID string, roleID string) error {
	return uc.repo.AssignPlatformAdminRole(ctx, adminID, roleID)
}

// ListPlatformAdminRoles lists assigned roles for a platform admin.
func (uc *IAMUsecase) ListPlatformAdminRoles(ctx context.Context, adminID string) ([]PlatformRole, error) {
	return uc.repo.ListPlatformAdminRoles(ctx, adminID)
}

// RemovePlatformAdminRole removes an assigned role from platform admin.
func (uc *IAMUsecase) RemovePlatformAdminRole(ctx context.Context, adminID string, roleID string) error {
	return uc.repo.RemovePlatformAdminRole(ctx, adminID, roleID)
}

// AssignPlatformRolePermissions assigns permissions to a platform role.
func (uc *IAMUsecase) AssignPlatformRolePermissions(ctx context.Context, roleID string, permissionCodes []string) error {
	return uc.repo.AssignPlatformRolePermissions(ctx, roleID, permissionCodes)
}

// ListPlatformRolePermissions lists permissions for a platform role.
func (uc *IAMUsecase) ListPlatformRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	return uc.repo.ListPlatformRolePermissions(ctx, roleID)
}

// CreateUser creates a tenant user.
func (uc *IAMUsecase) CreateUser(ctx context.Context, user User) (User, error) {
	return uc.repo.CreateUser(ctx, user)
}

// ListUsers lists users in current tenant.
func (uc *IAMUsecase) ListUsers(ctx context.Context) ([]User, error) {
	return uc.repo.ListUsers(ctx)
}

// CreateRole creates a tenant role.
func (uc *IAMUsecase) CreateRole(ctx context.Context, role Role) (Role, error) {
	return uc.repo.CreateRole(ctx, role)
}

// ListRoles lists roles in current tenant.
func (uc *IAMUsecase) ListRoles(ctx context.Context) ([]Role, error) {
	return uc.repo.ListRoles(ctx)
}

// AssignRole assigns role to user within tenant.
func (uc *IAMUsecase) AssignRole(ctx context.Context, userID string, roleID string) error {
	return uc.repo.AssignRole(ctx, userID, roleID)
}

// ListUserRoles lists roles for a user.
func (uc *IAMUsecase) ListUserRoles(ctx context.Context, userID string) ([]Role, error) {
	return uc.repo.ListUserRoles(ctx, userID)
}

// CreatePermission creates a permission (platform admin).
func (uc *IAMUsecase) CreatePermission(ctx context.Context, permission Permission) (Permission, error) {
	return uc.repo.CreatePermission(ctx, permission)
}

// ListPermissions lists permission catalog.
func (uc *IAMUsecase) ListPermissions(ctx context.Context) ([]Permission, error) {
	return uc.repo.ListPermissions(ctx)
}

// ListTenantPermissions lists permissions visible to tenants.
func (uc *IAMUsecase) ListTenantPermissions(ctx context.Context) ([]Permission, error) {
	return uc.repo.ListTenantPermissions(ctx)
}

// AssignRolePermissions assigns permissions to a role.
func (uc *IAMUsecase) AssignRolePermissions(ctx context.Context, roleID string, permissionCodes []string) error {
	return uc.repo.AssignRolePermissions(ctx, roleID, permissionCodes)
}

// ListRolePermissions lists permissions for a role.
func (uc *IAMUsecase) ListRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	return uc.repo.ListRolePermissions(ctx, roleID)
}

// RequirePermission enforces RBAC based on JWT subject.
func (uc *IAMUsecase) RequirePermission(ctx context.Context, permission string) error {
	if permission == "" {
		return nil
	}
	claims, ok := jwt.ClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return errors.Forbidden("RBAC_FORBIDDEN", "missing subject")
	}
	if strings.HasPrefix(permission, "platform.") {
		perms, err := uc.repo.ListPlatformAdminPermissions(ctx, claims.Subject)
		if err != nil {
			return err
		}
		for _, item := range perms {
			if permissionMatch(item.Code, permission) {
				return nil
			}
		}
		return errors.Forbidden("RBAC_FORBIDDEN", "permission denied")
	}
	if strings.HasPrefix(permission, "tenant.") {
		if hasRole(claims.Roles, "tenant_admin") {
			return nil
		}
		perms, err := uc.repo.ListUserPermissions(ctx, claims.Subject)
		if err != nil {
			return err
		}
		for _, item := range perms {
			if permissionMatch(item.Code, permission) {
				return nil
			}
		}
		return errors.Forbidden("RBAC_FORBIDDEN", "permission denied")
	}
	return errors.Forbidden("RBAC_FORBIDDEN", "invalid permission namespace")
}

func permissionMatch(candidate string, target string) bool {
	if candidate == "*" || candidate == target {
		return true
	}
	if strings.HasSuffix(candidate, ".*") {
		prefix := strings.TrimSuffix(candidate, ".*")
		return strings.HasPrefix(target, prefix+".")
	}
	return false
}

func hasRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

// ProviderSet is iam biz providers.
var ProviderSet = wire.NewSet(NewIAMUsecase)
