package data

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	internaldata "github.com/ZTH7/RagoDesk/apps/server/internal/data"
	biz "github.com/ZTH7/RagoDesk/apps/server/internal/iam/biz"
	"github.com/ZTH7/RagoDesk/apps/server/internal/kit/tenant"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/google/wire"
)

type iamRepo struct {
	log *log.Helper
	db  *sql.DB
}

// NewIAMRepo creates a new iam repo.
func NewIAMRepo(data *internaldata.Data, logger log.Logger) biz.IAMRepo {
	return &iamRepo{log: log.NewHelper(logger), db: data.DB}
}

func (r *iamRepo) Ping(ctx context.Context) error {
	if r.db == nil {
		return kerrors.InternalServer("DB_MISSING", "database not initialized")
	}
	return r.db.PingContext(ctx)
}

func (r *iamRepo) CreateTenant(ctx context.Context, tenantModel biz.Tenant) (biz.Tenant, error) {
	// TODO: platform admin auth & audit.
	if tenantModel.ID == "" {
		tenantModel.ID = uuid.NewString()
	}
	if tenantModel.Type == "" {
		tenantModel.Type = "enterprise"
	}
	if tenantModel.CreatedAt.IsZero() {
		tenantModel.CreatedAt = time.Now()
	}
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO tenant (id, name, type, plan, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		tenantModel.ID,
		tenantModel.Name,
		tenantModel.Type,
		tenantModel.Plan,
		tenantModel.Status,
		tenantModel.CreatedAt,
	)
	if err != nil {
		return biz.Tenant{}, err
	}
	return tenantModel, nil
}

func (r *iamRepo) GetTenant(ctx context.Context, id string) (biz.Tenant, error) {
	// TODO: platform admin auth & audit.
	var tenantModel biz.Tenant
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, name, type, plan, status, created_at FROM tenant WHERE id = ?",
		id,
	).Scan(&tenantModel.ID, &tenantModel.Name, &tenantModel.Type, &tenantModel.Plan, &tenantModel.Status, &tenantModel.CreatedAt)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return biz.Tenant{}, kerrors.NotFound("TENANT_NOT_FOUND", "tenant not found")
		}
		return biz.Tenant{}, err
	}
	return tenantModel, nil
}

func (r *iamRepo) ListTenants(ctx context.Context) ([]biz.Tenant, error) {
	// TODO: platform admin auth & audit.
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, type, plan, status, created_at FROM tenant ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Tenant, 0)
	for rows.Next() {
		var tenantModel biz.Tenant
		if err := rows.Scan(&tenantModel.ID, &tenantModel.Name, &tenantModel.Type, &tenantModel.Plan, &tenantModel.Status, &tenantModel.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, tenantModel)
	}
	return items, rows.Err()
}

func (r *iamRepo) CreatePlatformAdmin(ctx context.Context, admin biz.PlatformAdmin) (biz.PlatformAdmin, error) {
	if admin.ID == "" {
		admin.ID = uuid.NewString()
	}
	if admin.CreatedAt.IsZero() {
		admin.CreatedAt = time.Now()
	}
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO platform_admin (id, email, phone, name, status, password_hash, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		admin.ID,
		admin.Email,
		admin.Phone,
		admin.Name,
		admin.Status,
		admin.PasswordHash,
		admin.CreatedAt,
	)
	if err != nil {
		return biz.PlatformAdmin{}, err
	}
	return admin, nil
}

func (r *iamRepo) ListPlatformAdmins(ctx context.Context) ([]biz.PlatformAdmin, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, email, phone, name, status, created_at FROM platform_admin ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.PlatformAdmin, 0)
	for rows.Next() {
		var admin biz.PlatformAdmin
		if err := rows.Scan(&admin.ID, &admin.Email, &admin.Phone, &admin.Name, &admin.Status, &admin.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, admin)
	}
	return items, rows.Err()
}

func (r *iamRepo) CreatePlatformRole(ctx context.Context, role biz.PlatformRole) (biz.PlatformRole, error) {
	if role.ID == "" {
		role.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO platform_role (id, name) VALUES (?, ?)",
		role.ID,
		role.Name,
	)
	if err != nil {
		return biz.PlatformRole{}, err
	}
	return role, nil
}

func (r *iamRepo) ListPlatformRoles(ctx context.Context) ([]biz.PlatformRole, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM platform_role ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.PlatformRole, 0)
	for rows.Next() {
		var role biz.PlatformRole
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		items = append(items, role)
	}
	return items, rows.Err()
}

func (r *iamRepo) AssignPlatformAdminRole(ctx context.Context, adminID string, roleID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var adminExists string
	if err = tx.QueryRowContext(ctx, "SELECT id FROM platform_admin WHERE id = ?", adminID).Scan(&adminExists); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return kerrors.NotFound("PLATFORM_ADMIN_NOT_FOUND", "platform admin not found")
		}
		return err
	}
	var roleExists string
	if err = tx.QueryRowContext(ctx, "SELECT id FROM platform_role WHERE id = ?", roleID).Scan(&roleExists); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return kerrors.NotFound("PLATFORM_ROLE_NOT_FOUND", "platform role not found")
		}
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO platform_admin_role (admin_id, role_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE role_id = role_id",
		adminID,
		roleID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *iamRepo) ListPlatformAdminPermissions(ctx context.Context, adminID string) ([]biz.Permission, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT DISTINCT p.id, p.code, p.description, p.scope
		FROM platform_admin_role ar
		JOIN platform_role_permission rp ON ar.role_id = rp.role_id
		JOIN permission p ON rp.permission_id = p.id
		WHERE ar.admin_id = ? AND p.scope = ?`,
		adminID,
		biz.PermissionScopePlatform,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Permission, 0)
	for rows.Next() {
		var perm biz.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope); err != nil {
			return nil, err
		}
		items = append(items, perm)
	}
	return items, rows.Err()
}

func (r *iamRepo) AssignPlatformRolePermissions(ctx context.Context, roleID string, permissionCodes []string) error {
	if len(permissionCodes) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var roleExists string
	if err = tx.QueryRowContext(ctx, "SELECT id FROM platform_role WHERE id = ?", roleID).Scan(&roleExists); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return kerrors.NotFound("PLATFORM_ROLE_NOT_FOUND", "platform role not found")
		}
		return err
	}

	placeholders, args := buildInClause(permissionCodes)
	query := fmt.Sprintf("SELECT id, code, description, scope FROM permission WHERE scope = ? AND code IN (%s)", placeholders)
	args = append([]any{biz.PermissionScopePlatform}, args...)
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	permByCode := map[string]biz.Permission{}
	for rows.Next() {
		var perm biz.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope); err != nil {
			return err
		}
		permByCode[perm.Code] = perm
	}
	if err := rows.Err(); err != nil {
		return err
	}

	missing := missingPermissions(permissionCodes, permByCode)
	if len(missing) > 0 {
		return kerrors.NotFound("PERMISSION_NOT_FOUND", fmt.Sprintf("missing permissions: %s", strings.Join(missing, ",")))
	}

	for _, code := range permissionCodes {
		perm := permByCode[code]
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO platform_role_permission (role_id, permission_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE permission_id = permission_id",
			roleID,
			perm.ID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *iamRepo) ListPlatformRolePermissions(ctx context.Context, roleID string) ([]biz.Permission, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT p.id, p.code, p.description, p.scope
		FROM platform_role_permission rp
		JOIN permission p ON rp.permission_id = p.id
		WHERE rp.role_id = ? AND p.scope = ?`,
		roleID,
		biz.PermissionScopePlatform,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Permission, 0)
	for rows.Next() {
		var perm biz.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope); err != nil {
			return nil, err
		}
		items = append(items, perm)
	}
	return items, rows.Err()
}

func (r *iamRepo) CreateUser(ctx context.Context, user biz.User) (biz.User, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.User{}, err
	}
	user.TenantID = tenantID
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO `user` (id, tenant_id, email, phone, name, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.TenantID,
		user.Email,
		user.Phone,
		user.Name,
		user.Status,
		user.CreatedAt,
	)
	if err != nil {
		return biz.User{}, err
	}
	return user, nil
}

func (r *iamRepo) ListUsers(ctx context.Context) ([]biz.User, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, tenant_id, email, phone, name, status, created_at FROM `user` WHERE tenant_id = ? ORDER BY created_at DESC",
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.User, 0)
	for rows.Next() {
		var user biz.User
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.Phone, &user.Name, &user.Status, &user.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, user)
	}
	return items, rows.Err()
}

func (r *iamRepo) CreateRole(ctx context.Context, role biz.Role) (biz.Role, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.Role{}, err
	}
	role.TenantID = tenantID
	if role.ID == "" {
		role.ID = uuid.NewString()
	}
	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO `role` (id, tenant_id, name) VALUES (?, ?, ?)",
		role.ID,
		role.TenantID,
		role.Name,
	)
	if err != nil {
		return biz.Role{}, err
	}
	return role, nil
}

func (r *iamRepo) ListRoles(ctx context.Context) ([]biz.Role, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, tenant_id, name FROM `role` WHERE tenant_id = ? ORDER BY name ASC",
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Role, 0)
	for rows.Next() {
		var role biz.Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name); err != nil {
			return nil, err
		}
		items = append(items, role)
	}
	return items, rows.Err()
}

func (r *iamRepo) AssignRole(ctx context.Context, userID string, roleID string) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var userTenant string
	if err = tx.QueryRowContext(ctx, "SELECT tenant_id FROM `user` WHERE id = ?", userID).Scan(&userTenant); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return kerrors.NotFound("USER_NOT_FOUND", "user not found")
		}
		return err
	}
	if userTenant != tenantID {
		return kerrors.Forbidden("TENANT_MISMATCH", "tenant mismatch")
	}

	var roleTenant string
	if err = tx.QueryRowContext(ctx, "SELECT tenant_id FROM `role` WHERE id = ?", roleID).Scan(&roleTenant); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return kerrors.NotFound("ROLE_NOT_FOUND", "role not found")
		}
		return err
	}
	if roleTenant != tenantID {
		return kerrors.Forbidden("TENANT_MISMATCH", "tenant mismatch")
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO user_role (user_id, role_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE role_id = role_id",
		userID,
		roleID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *iamRepo) ListUserRoles(ctx context.Context, userID string) ([]biz.Role, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT r.id, r.tenant_id, r.name
		FROM user_role ur
		JOIN `+"`role`"+` r ON ur.role_id = r.id
		JOIN `+"`user`"+` u ON ur.user_id = u.id
		WHERE u.id = ? AND u.tenant_id = ? AND r.tenant_id = u.tenant_id`,
		userID,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Role, 0)
	for rows.Next() {
		var role biz.Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name); err != nil {
			return nil, err
		}
		items = append(items, role)
	}
	return items, rows.Err()
}

func (r *iamRepo) ListUserPermissions(ctx context.Context, userID string) ([]biz.Permission, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT DISTINCT p.id, p.code, p.description, p.scope
		FROM `+"`user`"+` u
		JOIN user_role ur ON u.id = ur.user_id
		JOIN role_permission rp ON ur.role_id = rp.role_id
		JOIN permission p ON rp.permission_id = p.id
		WHERE u.id = ? AND u.tenant_id = ? AND p.scope = ?`,
		userID,
		tenantID,
		biz.PermissionScopeTenant,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Permission, 0)
	for rows.Next() {
		var perm biz.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope); err != nil {
			return nil, err
		}
		items = append(items, perm)
	}
	return items, rows.Err()
}

func (r *iamRepo) CreatePermission(ctx context.Context, permission biz.Permission) (biz.Permission, error) {
	if permission.ID == "" {
		permission.ID = uuid.NewString()
	}
	if permission.Scope == "" {
		permission.Scope = biz.PermissionScopePlatform
	}
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO permission (id, code, description, scope) VALUES (?, ?, ?, ?)",
		permission.ID,
		permission.Code,
		permission.Description,
		permission.Scope,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return r.getPermissionByCode(ctx, permission.Code)
		}
		return biz.Permission{}, err
	}
	return permission, nil
}

func (r *iamRepo) ListPermissions(ctx context.Context) ([]biz.Permission, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, code, description, scope FROM permission ORDER BY code ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Permission, 0)
	for rows.Next() {
		var perm biz.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope); err != nil {
			return nil, err
		}
		items = append(items, perm)
	}
	return items, rows.Err()
}

func (r *iamRepo) ListTenantPermissions(ctx context.Context) ([]biz.Permission, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, code, description, scope FROM permission WHERE scope = ? ORDER BY code ASC",
		biz.PermissionScopeTenant,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Permission, 0)
	for rows.Next() {
		var perm biz.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope); err != nil {
			return nil, err
		}
		items = append(items, perm)
	}
	return items, rows.Err()
}

func (r *iamRepo) AssignRolePermissions(ctx context.Context, roleID string, permissionCodes []string) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	if len(permissionCodes) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var roleTenant string
	if err = tx.QueryRowContext(ctx, "SELECT tenant_id FROM `role` WHERE id = ?", roleID).Scan(&roleTenant); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return kerrors.NotFound("ROLE_NOT_FOUND", "role not found")
		}
		return err
	}
	if roleTenant != tenantID {
		return kerrors.Forbidden("TENANT_MISMATCH", "tenant mismatch")
	}

	placeholders, args := buildInClause(permissionCodes)
	query := fmt.Sprintf("SELECT id, code, description, scope FROM permission WHERE scope = ? AND code IN (%s)", placeholders)
	args = append([]any{biz.PermissionScopeTenant}, args...)
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	permByCode := map[string]biz.Permission{}
	for rows.Next() {
		var perm biz.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope); err != nil {
			return err
		}
		permByCode[perm.Code] = perm
	}
	if err := rows.Err(); err != nil {
		return err
	}

	missing := missingPermissions(permissionCodes, permByCode)
	if len(missing) > 0 {
		return kerrors.NotFound("PERMISSION_NOT_FOUND", fmt.Sprintf("missing permissions: %s", strings.Join(missing, ",")))
	}

	for _, code := range permissionCodes {
		perm := permByCode[code]
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO role_permission (role_id, permission_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE permission_id = permission_id",
			roleID,
			perm.ID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *iamRepo) ListRolePermissions(ctx context.Context, roleID string) ([]biz.Permission, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT p.id, p.code, p.description, p.scope
		FROM role_permission rp
		JOIN permission p ON rp.permission_id = p.id
		JOIN `+"`role`"+` r ON rp.role_id = r.id
		WHERE r.id = ? AND r.tenant_id = ? AND p.scope = ?`,
		roleID,
		tenantID,
		biz.PermissionScopeTenant,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]biz.Permission, 0)
	for rows.Next() {
		var perm biz.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope); err != nil {
			return nil, err
		}
		items = append(items, perm)
	}
	return items, rows.Err()
}

func (r *iamRepo) getPermissionByCode(ctx context.Context, code string) (biz.Permission, error) {
	var perm biz.Permission
	err := r.db.QueryRowContext(ctx, "SELECT id, code, description, scope FROM permission WHERE code = ?", code).
		Scan(&perm.ID, &perm.Code, &perm.Description, &perm.Scope)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return biz.Permission{}, kerrors.NotFound("PERMISSION_NOT_FOUND", "permission not found")
		}
		return biz.Permission{}, err
	}
	return perm, nil
}

func buildInClause(values []string) (string, []any) {
	placeholders := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		placeholders = append(placeholders, "?")
		args = append(args, value)
	}
	if len(placeholders) == 0 {
		placeholders = append(placeholders, "?")
		args = append(args, "")
	}
	return strings.Join(placeholders, ","), args
}

func missingPermissions(codes []string, existing map[string]biz.Permission) []string {
	seen := map[string]struct{}{}
	missing := make([]string, 0)
	for _, code := range codes {
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		if _, ok := existing[code]; !ok {
			missing = append(missing, code)
		}
	}
	return missing
}

// ProviderSet is iam data providers.
var ProviderSet = wire.NewSet(NewIAMRepo)
