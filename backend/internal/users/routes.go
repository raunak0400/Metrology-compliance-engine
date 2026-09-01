package users

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

	g := rg.Group("/users")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.GET("", middleware.Authorize(constants.RoleAdmin, constants.RoleSupervisor), func(c *gin.Context) {
		var q ListUsersQuery
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

	g.GET("/:id", func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		user, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, user)
	})

	g.PATCH("/:id", middleware.Authorize(constants.RoleAdmin), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req UpdateUserRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		before, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		user, err := svc.Update(c.Request.Context(), id, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionUPDATE, EntityType: "User", EntityID: id.String(), OldState: before, NewState: user})
		response.OK(c, user)
	})
}
