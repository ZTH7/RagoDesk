package biz

import (
	"context"
	"strings"
	"time"

	"github.com/ZTH7/RagoDesk/apps/server/internal/auth"
	"github.com/ZTH7/RagoDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/google/wire"
	"golang.org/x/crypto/bcrypt"
)

const defaultTokenTTL = 24 * time.Hour

// Tenant represents a tenant for registration.
type Tenant struct {
	ID        string
	Name      string
	Type      string
	Plan      string
	Status    string
	CreatedAt time.Time
}

// TenantAccount represents a tenant user for authentication.
type TenantAccount struct {
	ID           string
	TenantID     string
	Email        string
	Phone        string
	Name         string
	Status       string
	PasswordHash string
}

// PlatformAccount represents a platform admin for authentication.
type PlatformAccount struct {
	ID           string
	Email        string
	Phone        string
	Name         string
	Status       string
	PasswordHash string
}

// AuthProfile is the response profile for auth.
type AuthProfile struct {
	SubjectID string
	TenantID  string
	Account   string
	Name      string
	Roles     []string
}

// AuthSession holds authentication result.
type AuthSession struct {
	Token     string
	ExpiresAt time.Time
	Profile   AuthProfile
}

// ConsoleRegisterInput captures self-service registration input.
type ConsoleRegisterInput struct {
	TenantName string
	TenantType string
	AdminName  string
	Email      string
	Phone      string
	Password   string
}

// AuthRepo defines required auth repository methods.
type AuthRepo interface {
	FindTenantAccount(ctx context.Context, account string, tenantID string) (TenantAccount, error)
	FindPlatformAccount(ctx context.Context, account string) (PlatformAccount, error)
	ListUserRoles(ctx context.Context, userID string) ([]string, error)
	CreateTenantWithAdmin(ctx context.Context, tenant Tenant, admin TenantAccount, roleName string) (Tenant, TenantAccount, error)
}

// AuthUsecase handles authentication logic.
type AuthUsecase struct {
	repo      AuthRepo
	log       *log.Helper
	secret    string
	issuer    string
	audience  string
	tokenTTL  time.Duration
}

// NewAuthUsecase creates a new AuthUsecase.
func NewAuthUsecase(repo AuthRepo, cfg *conf.Server, logger log.Logger) *AuthUsecase {
	uc := &AuthUsecase{repo: repo, log: log.NewHelper(logger), tokenTTL: defaultTokenTTL}
	if cfg != nil && cfg.Auth != nil {
		uc.secret = strings.TrimSpace(cfg.Auth.JwtSecret)
		uc.issuer = strings.TrimSpace(cfg.Auth.Issuer)
		uc.audience = strings.TrimSpace(cfg.Auth.Audience)
	}
	return uc
}

func (uc *AuthUsecase) ConsoleLogin(ctx context.Context, account string, password string, tenantID string) (AuthSession, error) {
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	if account == "" || password == "" {
		return AuthSession{}, errors.BadRequest("LOGIN_INVALID", "account and password required")
	}
	user, err := uc.repo.FindTenantAccount(ctx, account, strings.TrimSpace(tenantID))
	if err != nil {
		if kerr, ok := err.(*errors.Error); ok && kerr.Reason == "ACCOUNT_AMBIGUOUS" {
			return AuthSession{}, err
		}
		return AuthSession{}, errors.Unauthorized("LOGIN_FAILED", "invalid credentials")
	}
	if !isActiveStatus(user.Status) {
		return AuthSession{}, errors.Forbidden("ACCOUNT_DISABLED", "account disabled")
	}
	if user.PasswordHash == "" {
		return AuthSession{}, errors.Unauthorized("LOGIN_FAILED", "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthSession{}, errors.Unauthorized("LOGIN_FAILED", "invalid credentials")
	}
	roles, err := uc.repo.ListUserRoles(ctx, user.ID)
	if err != nil {
		return AuthSession{}, err
	}
	profile := AuthProfile{
		SubjectID: user.ID,
		TenantID:  user.TenantID,
		Account:   primaryAccount(user.Email, user.Phone),
		Name:      user.Name,
		Roles:     roles,
	}
	return uc.buildSession(profile)
}

func (uc *AuthUsecase) ConsoleRegister(ctx context.Context, input ConsoleRegisterInput) (AuthSession, error) {
	input.TenantName = strings.TrimSpace(input.TenantName)
	input.AdminName = strings.TrimSpace(input.AdminName)
	input.Email = strings.TrimSpace(input.Email)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Password = strings.TrimSpace(input.Password)
	if input.TenantName == "" || input.Password == "" || (input.Email == "" && input.Phone == "") {
		return AuthSession{}, errors.BadRequest("REGISTER_INVALID", "tenant, account, and password required")
	}
	tenantType := strings.TrimSpace(input.TenantType)
	if tenantType == "" {
		tenantType = "enterprise"
	}
	if tenantType != "enterprise" && tenantType != "personal" {
		return AuthSession{}, errors.BadRequest("TENANT_TYPE_INVALID", "invalid tenant type")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthSession{}, errors.InternalServer("PASSWORD_HASH_FAILED", "password hash failed")
	}
	now := time.Now()
	tenant := Tenant{
		ID:        uuid.NewString(),
		Name:      input.TenantName,
		Type:      tenantType,
		Plan:      "free",
		Status:    "active",
		CreatedAt: now,
	}
	admin := TenantAccount{
		ID:           uuid.NewString(),
		TenantID:     tenant.ID,
		Email:        input.Email,
		Phone:        input.Phone,
		Name:         input.AdminName,
		Status:       "active",
		PasswordHash: string(passwordHash),
	}
	createdTenant, createdAdmin, err := uc.repo.CreateTenantWithAdmin(ctx, tenant, admin, "tenant_admin")
	if err != nil {
		return AuthSession{}, err
	}
	roles, err := uc.repo.ListUserRoles(ctx, createdAdmin.ID)
	if err != nil {
		return AuthSession{}, err
	}
	profile := AuthProfile{
		SubjectID: createdAdmin.ID,
		TenantID:  createdTenant.ID,
		Account:   primaryAccount(createdAdmin.Email, createdAdmin.Phone),
		Name:      createdAdmin.Name,
		Roles:     roles,
	}
	return uc.buildSession(profile)
}

func (uc *AuthUsecase) PlatformLogin(ctx context.Context, account string, password string) (AuthSession, error) {
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	if account == "" || password == "" {
		return AuthSession{}, errors.BadRequest("LOGIN_INVALID", "account and password required")
	}
	admin, err := uc.repo.FindPlatformAccount(ctx, account)
	if err != nil {
		return AuthSession{}, errors.Unauthorized("LOGIN_FAILED", "invalid credentials")
	}
	if !isActiveStatus(admin.Status) {
		return AuthSession{}, errors.Forbidden("ACCOUNT_DISABLED", "account disabled")
	}
	if admin.PasswordHash == "" {
		return AuthSession{}, errors.Unauthorized("LOGIN_FAILED", "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return AuthSession{}, errors.Unauthorized("LOGIN_FAILED", "invalid credentials")
	}
	profile := AuthProfile{
		SubjectID: admin.ID,
		Account:   primaryAccount(admin.Email, admin.Phone),
		Name:      admin.Name,
		Roles:     []string{"platform_admin"},
	}
	return uc.buildSession(profile)
}

func (uc *AuthUsecase) buildSession(profile AuthProfile) (AuthSession, error) {
	if strings.TrimSpace(uc.secret) == "" {
		return AuthSession{}, errors.InternalServer("JWT_SECRET_MISSING", "jwt secret missing")
	}
	now := time.Now()
	expires := now.Add(uc.tokenTTL)
	claims := auth.Claims{
		TenantID:  profile.TenantID,
		Subject:   profile.SubjectID,
		Issuer:    uc.issuer,
		Expiry:    expires.Unix(),
		NotBefore: now.Unix(),
		IssuedAt:  now.Unix(),
		Roles:     profile.Roles,
	}
	if uc.audience != "" {
		claims.Audience = uc.audience
	}
	token, err := auth.SignHS256(claims, uc.secret)
	if err != nil {
		return AuthSession{}, errors.InternalServer("JWT_SIGN_FAILED", "sign token failed")
	}
	return AuthSession{
		Token:     token,
		ExpiresAt: expires,
		Profile:   profile,
	}, nil
}

func primaryAccount(email string, phone string) string {
	if strings.TrimSpace(email) != "" {
		return email
	}
	return phone
}

func isActiveStatus(status string) bool {
	return strings.TrimSpace(strings.ToLower(status)) == "active"
}

// ProviderSet is authn biz providers.
var ProviderSet = wire.NewSet(NewAuthUsecase)
