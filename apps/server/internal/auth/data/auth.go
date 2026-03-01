package data

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	biz "github.com/ZTH7/RagoDesk/apps/server/internal/auth/biz"
	internaldata "github.com/ZTH7/RagoDesk/apps/server/internal/data"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/google/wire"
)

type authRepo struct {
	log *log.Helper
	db  *sql.DB
}

// NewAuthRepo creates a new auth repo.
func NewAuthRepo(data *internaldata.Data, logger log.Logger) biz.AuthRepo {
	return &authRepo{log: log.NewHelper(logger), db: data.DB}
}

func (r *authRepo) FindTenantAccount(ctx context.Context, account string, tenantID string) (biz.TenantAccount, error) {
	account = strings.TrimSpace(account)
	if account == "" {
		return biz.TenantAccount{}, kerrors.BadRequest("ACCOUNT_REQUIRED", "account required")
	}
	if strings.TrimSpace(tenantID) != "" {
		var user biz.TenantAccount
		err := r.db.QueryRowContext(
			ctx,
			"SELECT id, tenant_id, email, phone, name, status, password_hash FROM `user` WHERE tenant_id = ? AND (email = ? OR phone = ?)",
			tenantID,
			account,
			account,
		).Scan(&user.ID, &user.TenantID, &user.Email, &user.Phone, &user.Name, &user.Status, &user.PasswordHash)
		if err != nil {
			if stderrors.Is(err, sql.ErrNoRows) {
				return biz.TenantAccount{}, kerrors.NotFound("ACCOUNT_NOT_FOUND", "account not found")
			}
			return biz.TenantAccount{}, err
		}
		return user, nil
	}
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, tenant_id, email, phone, name, status, password_hash FROM `user` WHERE email = ? OR phone = ? ORDER BY created_at DESC LIMIT 2",
		account,
		account,
	)
	if err != nil {
		return biz.TenantAccount{}, err
	}
	defer rows.Close()
	items := make([]biz.TenantAccount, 0, 2)
	for rows.Next() {
		var user biz.TenantAccount
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.Phone, &user.Name, &user.Status, &user.PasswordHash); err != nil {
			return biz.TenantAccount{}, err
		}
		items = append(items, user)
	}
	if err := rows.Err(); err != nil {
		return biz.TenantAccount{}, err
	}
	if len(items) == 0 {
		return biz.TenantAccount{}, kerrors.NotFound("ACCOUNT_NOT_FOUND", "account not found")
	}
	if len(items) > 1 {
		return biz.TenantAccount{}, kerrors.BadRequest("ACCOUNT_AMBIGUOUS", "account matches multiple tenants")
	}
	return items[0], nil
}

func (r *authRepo) FindPlatformAccount(ctx context.Context, account string) (biz.PlatformAccount, error) {
	account = strings.TrimSpace(account)
	if account == "" {
		return biz.PlatformAccount{}, kerrors.BadRequest("ACCOUNT_REQUIRED", "account required")
	}
	var admin biz.PlatformAccount
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, email, phone, name, status, password_hash FROM platform_admin WHERE email = ? OR phone = ?",
		account,
		account,
	).Scan(&admin.ID, &admin.Email, &admin.Phone, &admin.Name, &admin.Status, &admin.PasswordHash)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return biz.PlatformAccount{}, kerrors.NotFound("ACCOUNT_NOT_FOUND", "account not found")
		}
		return biz.PlatformAccount{}, err
	}
	return admin, nil
}

func (r *authRepo) ListUserRoles(ctx context.Context, userID string) ([]string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT r.name
		FROM user_role ur
		JOIN `+"`role`"+` r ON ur.role_id = r.id
		JOIN `+"`user`"+` u ON ur.user_id = u.id
		WHERE u.id = ? AND r.tenant_id = u.tenant_id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		items = append(items, name)
	}
	return items, rows.Err()
}

func (r *authRepo) CreateTenantWithAdmin(ctx context.Context, tenant biz.Tenant, admin biz.TenantAccount, roleName string) (biz.Tenant, biz.TenantAccount, error) {
	if roleName == "" {
		roleName = "tenant_admin"
	}
	if existingTenantID, err := r.findExistingTenantByAccount(ctx, admin.Email, admin.Phone); err != nil {
		return biz.Tenant{}, biz.TenantAccount{}, err
	} else if existingTenantID != "" {
		return biz.Tenant{}, biz.TenantAccount{}, kerrors.Conflict(
			"ACCOUNT_ALREADY_BOUND",
			fmt.Sprintf("account already bound to tenant %s", existingTenantID),
		)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return biz.Tenant{}, biz.TenantAccount{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO tenant (id, name, type, plan, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		tenant.ID,
		tenant.Name,
		tenant.Type,
		tenant.Plan,
		tenant.Status,
		tenant.CreatedAt,
	)
	if err != nil {
		return biz.Tenant{}, biz.TenantAccount{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO `user` (id, tenant_id, email, phone, name, status, password_hash, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		admin.ID,
		tenant.ID,
		emptyToNull(admin.Email),
		emptyToNull(admin.Phone),
		emptyToNull(admin.Name),
		admin.Status,
		emptyToNull(admin.PasswordHash),
		time.Now(),
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return biz.Tenant{}, biz.TenantAccount{}, kerrors.Conflict("ACCOUNT_EXISTS", "account already exists")
		}
		return biz.Tenant{}, biz.TenantAccount{}, err
	}

	roleUUID := uuid.NewString()
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO `role` (id, tenant_id, name) VALUES (?, ?, ?)",
		roleUUID,
		tenant.ID,
		roleName,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return biz.Tenant{}, biz.TenantAccount{}, kerrors.Conflict("ROLE_EXISTS", "role already exists")
		}
		return biz.Tenant{}, biz.TenantAccount{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO user_role (user_id, role_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE role_id = role_id",
		admin.ID,
		roleUUID,
	)
	if err != nil {
		return biz.Tenant{}, biz.TenantAccount{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO role_permission (role_id, permission_id) SELECT ?, id FROM permission WHERE scope = ?",
		roleUUID,
		"tenant",
	)
	if err != nil {
		return biz.Tenant{}, biz.TenantAccount{}, err
	}

	if err = tx.Commit(); err != nil {
		return biz.Tenant{}, biz.TenantAccount{}, err
	}
	return tenant, admin, nil
}

func emptyToNull(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func (r *authRepo) findExistingTenantByAccount(ctx context.Context, email string, phone string) (string, error) {
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)
	if email == "" && phone == "" {
		return "", nil
	}
	query := "SELECT tenant_id FROM `user` WHERE "
	args := make([]any, 0, 2)
	if email != "" && phone != "" {
		query += "(email = ? OR phone = ?)"
		args = append(args, email, phone)
	} else if email != "" {
		query += "email = ?"
		args = append(args, email)
	} else {
		query += "phone = ?"
		args = append(args, phone)
	}
	query += " LIMIT 1"

	var tenantID string
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&tenantID); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return tenantID, nil
}

// ProviderSet is auth data providers.
var ProviderSet = wire.NewSet(NewAuthRepo)
