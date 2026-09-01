// Authentication + RBAC middleware.
//
// Authenticate verifies the JWT access token and attaches the authenticated
// user to the Gin context. Authorization is expressed via reusable policy
// middleware (Authorize, RequirePermission) — handlers must NOT hard-code role
// checks inline.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/authtoken"
)

const userContextKey = "authUser"

// AuthUser is the authenticated principal attached to the request context.
type AuthUser struct {
	ID          string
	Email       string
	Roles       []string
	Permissions []string
}

func (u AuthUser) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// CurrentUser reads the authenticated user set by Authenticate/OptionalAuth.
func CurrentUser(c *gin.Context) (AuthUser, bool) {
	v, ok := c.Get(userContextKey)
	if !ok {
		return AuthUser{}, false
	}
	u, ok := v.(AuthUser)
	return u, ok
}

func extractBearer(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

// Authenticate requires a valid access token and populates the request's AuthUser.
func Authenticate(tokenizer *authtoken.Tokenizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c)
		if token == "" {
			_ = c.Error(apperror.Unauthorized("Missing bearer token"))
			c.Abort()
			return
		}
		claims, err := tokenizer.VerifyAccessToken(token)
		if err != nil {
			_ = c.Error(apperror.Unauthorized("Invalid or expired token"))
			c.Abort()
			return
		}
		c.Set(userContextKey, AuthUser{
			ID:          claims.Subject,
			Email:       claims.Email,
			Roles:       orEmpty(claims.Roles),
			Permissions: orEmpty(claims.Permissions),
		})
		c.Next()
	}
}

// OptionalAuth attaches AuthUser if a valid token is present, but does not require it.
func OptionalAuth(tokenizer *authtoken.Tokenizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c)
		if token == "" {
			c.Next()
			return
		}
		claims, err := tokenizer.VerifyAccessToken(token)
		if err == nil {
			c.Set(userContextKey, AuthUser{
				ID:          claims.Subject,
				Email:       claims.Email,
				Roles:       orEmpty(claims.Roles),
				Permissions: orEmpty(claims.Permissions),
			})
		}
		c.Next()
	}
}

// Authorize requires the user to hold at least one of the given roles (ADMIN always passes).
func Authorize(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			_ = c.Error(apperror.Unauthorized(""))
			c.Abort()
			return
		}
		if user.HasRole("ADMIN") {
			c.Next()
			return
		}
		for _, r := range roles {
			if user.HasRole(r) {
				c.Next()
				return
			}
		}
		_ = c.Error(apperror.Forbidden("Requires one of roles: " + strings.Join(roles, ", ")))
		c.Abort()
	}
}

// RequirePermission requires the user to hold all listed permission scopes
// ("*" or the ADMIN role grants everything).
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			_ = c.Error(apperror.Unauthorized(""))
			c.Abort()
			return
		}
		if user.HasRole("ADMIN") || contains(user.Permissions, "*") {
			c.Next()
			return
		}
		var missing []string
		for _, p := range permissions {
			if !contains(user.Permissions, p) {
				missing = append(missing, p)
			}
		}
		if len(missing) > 0 {
			_ = c.Error(apperror.Forbidden("Missing permission(s): " + strings.Join(missing, ", ")))
			c.Abort()
			return
		}
		c.Next()
	}
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
