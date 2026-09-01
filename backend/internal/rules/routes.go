package rules

import (
	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/audit"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/httpapi"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/pagination"
	"github.com/raunak0400/metrology-backend/internal/response"
)

func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps) {
	svc := NewService(d.Queries, d.Pool)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/rules")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.GET("", middleware.RequirePermission(constants.PermRulesRead), func(c *gin.Context) {
		var q ListRulesQuery
		if err := httpapi.BindQuery(c, &q); err != nil {
			_ = c.Error(err)
			return
		}
		page := pagination.Parse(c)
		items, total, err := svc.List(c.Request.Context(), q, page)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OKWithMeta(c, items, pagination.Build(items, total, page))
	})

	g.GET("/:id", middleware.RequirePermission(constants.PermRulesRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		rule, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, rule)
	})

	g.POST("", middleware.RequirePermission(constants.PermRulesWrite), func(c *gin.Context) {
		var req CreateRuleRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		user, _ := middleware.CurrentUser(c)
		rule, err := svc.Create(c.Request.Context(), req, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionCREATE, EntityType: "LegalRule", EntityID: rule.ID.String(), NewState: rule})
		response.Created(c, rule)
	})

	g.PATCH("/:id", middleware.RequirePermission(constants.PermRulesWrite), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req UpdateRuleRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		rule, err := svc.Update(c.Request.Context(), id, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionUPDATE, EntityType: "LegalRule", EntityID: id.String(), NewState: rule})
		response.OK(c, rule)
	})
}
