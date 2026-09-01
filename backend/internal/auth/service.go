// Authentication service: registration, login, refresh-token rotation, logout.
// Refresh tokens are persisted as SHA-256 hashes with a jti so they can be
// revoked and rotated. Access tokens remain stateless.
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/authtoken"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
)

type Service struct {
	q         *db.Queries
	tokenizer *authtoken.Tokenizer
}

func NewService(q *db.Queries, tokenizer *authtoken.Tokenizer) *Service {
	return &Service{q: q, tokenizer: tokenizer}
}

type RegisteredUser struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"fullName"`
}

type Ctx struct {
	UserAgent string
	IP        string
}

type Tokens struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	Roles        []string `json:"roles"`
	Permissions  []string `json:"permissions"`
}

type LoginResult struct {
	User struct {
		ID       uuid.UUID `json:"id"`
		Email    string    `json:"email"`
		FullName string    `json:"fullName"`
		Roles    []string  `json:"roles"`
	} `json:"user"`
	Tokens
}

func (s *Service) rolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, []string, error) {
	rows, err := s.q.ListRolesForUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	roles := make([]string, 0, len(rows))
	permSet := map[string]struct{}{}
	for _, r := range rows {
		roles = append(roles, r.Name)
		for _, p := range r.Permissions {
			permSet[p] = struct{}{}
		}
	}
	perms := make([]string, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	return roles, perms, nil
}

func (s *Service) issueTokens(ctx context.Context, userID uuid.UUID, email string, actx Ctx) (Tokens, error) {
	roles, perms, err := s.rolesAndPermissions(ctx, userID)
	if err != nil {
		return Tokens{}, err
	}

	accessToken, err := s.tokenizer.SignAccessToken(userID.String(), email, roles, perms)
	if err != nil {
		return Tokens{}, err
	}

	jti := newJTI()
	refreshToken, expiresAt, err := s.tokenizer.SignRefreshToken(userID.String(), jti)
	if err != nil {
		return Tokens{}, err
	}

	var userAgent, ip *string
	if actx.UserAgent != "" {
		userAgent = &actx.UserAgent
	}
	if actx.IP != "" {
		ip = &actx.IP
	}

	tokenHash := sha256Hex(refreshToken)
	if _, err := s.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID: userID, Jti: jti, TokenHash: tokenHash, ExpiresAt: expiresAt, UserAgent: userAgent, Ip: ip,
	}); err != nil {
		return Tokens{}, err
	}

	return Tokens{AccessToken: accessToken, RefreshToken: refreshToken, Roles: roles, Permissions: perms}, nil
}

type RegisterInput struct {
	Email    string
	Password string
	FullName string
	Phone    string
}

// Register creates a user. New users default to the INSPECTOR role — admins
// adjust roles afterwards via PATCH /users/:id.
func (s *Service) Register(ctx context.Context, in RegisterInput) (RegisteredUser, error) {
	if _, err := s.q.GetUserByEmail(ctx, in.Email); err == nil {
		return RegisteredUser{}, apperror.Conflict("A user with this email already exists")
	}

	passwordHash, err := HashPassword(in.Password)
	if err != nil {
		return RegisteredUser{}, apperror.Internal("Could not hash password")
	}

	var phone *string
	if in.Phone != "" {
		phone = &in.Phone
	}

	user, err := s.q.CreateUser(ctx, db.CreateUserParams{
		Email: in.Email, PasswordHash: passwordHash, FullName: in.FullName, Phone: phone,
	})
	if err != nil {
		return RegisteredUser{}, err
	}

	if inspectorRole, err := s.q.GetRoleByName(ctx, constants.RoleInspector); err == nil {
		_ = s.q.InsertUserRole(ctx, db.InsertUserRoleParams{UserID: user.ID, RoleID: inspectorRole.ID})
	}

	return RegisteredUser{ID: user.ID, Email: user.Email, FullName: user.FullName}, nil
}

func (s *Service) Login(ctx context.Context, email, password string, actx Ctx) (LoginResult, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil || !user.IsActive {
		return LoginResult{}, apperror.Unauthorized("Invalid credentials")
	}

	ok, err := VerifyPassword(user.PasswordHash, password)
	if err != nil || !ok {
		return LoginResult{}, apperror.Unauthorized("Invalid credentials")
	}

	_ = s.q.SetLastLogin(ctx, user.ID)

	tokens, err := s.issueTokens(ctx, user.ID, user.Email, actx)
	if err != nil {
		return LoginResult{}, err
	}

	var result LoginResult
	result.User.ID = user.ID
	result.User.Email = user.Email
	result.User.FullName = user.FullName
	result.User.Roles = tokens.Roles
	result.Tokens = tokens
	return result, nil
}

func (s *Service) Refresh(ctx context.Context, token string, actx Ctx) (Tokens, error) {
	claims, err := s.tokenizer.VerifyRefreshToken(token)
	if err != nil {
		return Tokens{}, apperror.Unauthorized("Invalid refresh token")
	}

	stored, err := s.q.GetRefreshTokenByJTI(ctx, claims.JTI)
	if err != nil {
		return Tokens{}, apperror.Unauthorized("Refresh token is no longer valid")
	}
	if stored.RevokedAt != nil || stored.ExpiresAt.Before(time.Now()) {
		return Tokens{}, apperror.Unauthorized("Refresh token is no longer valid")
	}
	if stored.TokenHash != sha256Hex(token) {
		// Token reuse / mismatch — revoke defensively.
		_ = s.q.RevokeRefreshToken(ctx, claims.JTI)
		return Tokens{}, apperror.Unauthorized("Refresh token mismatch")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return Tokens{}, apperror.Unauthorized("Invalid refresh token")
	}
	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil || !user.IsActive {
		return Tokens{}, apperror.Unauthorized("User no longer active")
	}

	// Rotate: revoke old, issue new.
	_ = s.q.RevokeRefreshToken(ctx, claims.JTI)
	return s.issueTokens(ctx, user.ID, user.Email, actx)
}

// Logout is idempotent: invalid tokens are silently ignored.
func (s *Service) Logout(ctx context.Context, token string) {
	claims, err := s.tokenizer.VerifyRefreshToken(token)
	if err != nil {
		return
	}
	_ = s.q.RevokeRefreshTokenIfActive(ctx, claims.JTI)
}
