package roles

import (
	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/audit"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/httpapi"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/response"
)

func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps) {
	svc := NewService(d.Queries)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/roles")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.GET("", func(c *gin.Context) {
		items, err := svc.List(c.Request.Context())
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, items)
	})

	g.POST("", middleware.Authorize(constants.RoleAdmin), func(c *gin.Context) {
		var req CreateRoleRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		role, err := svc.Create(c.Request.Context(), req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionCREATE, EntityType: "Role", EntityID: role.ID.String(), NewState: role})
		response.Created(c, role)
	})
}
