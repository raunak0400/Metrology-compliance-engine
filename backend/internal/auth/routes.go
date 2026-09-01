package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/audit"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/httpapi"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/response"
)

func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps) {
	svc := NewService(d.Queries, d.Tokenizer)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/auth")
	// Stricter limiter for auth endpoints (brute-force protection).
	g.Use(middleware.RateLimit(d.Redis, middleware.RateLimiterConfig{Window: 15 * time.Minute, Max: 20, Prefix: "auth"}))

	g.POST("/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		user, err := svc.Register(c.Request.Context(), RegisterInput{
			Email: req.Email, Password: req.Password, FullName: req.FullName, Phone: req.Phone,
		})
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionCREATE, EntityType: "User", EntityID: user.ID.String(), NewState: user})
		response.Created(c, user)
	})

	g.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		result, err := svc.Login(c.Request.Context(), req.Email, req.Password, ctxOf(c))
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionLOGIN, EntityType: "User", EntityID: result.User.ID.String()})
		response.OK(c, result)
	})

	g.POST("/refresh", func(c *gin.Context) {
		var req RefreshRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		tokens, err := svc.Refresh(c.Request.Context(), req.RefreshToken, ctxOf(c))
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, tokens)
	})

	g.POST("/logout", func(c *gin.Context) {
		var req LogoutRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		svc.Logout(c.Request.Context(), req.RefreshToken)
		response.OK(c, gin.H{"message": "Logged out"})
	})
}

func ctxOf(c *gin.Context) Ctx {
	return Ctx{UserAgent: c.GetHeader("User-Agent"), IP: c.ClientIP()}
}
