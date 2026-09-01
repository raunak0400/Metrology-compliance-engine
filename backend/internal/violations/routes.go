package violations

import (
	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/audit"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/httpapi"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/response"
)

// RegisterRoutes owns /violations/:id (mounted at the API root).
func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps, inspSvc *inspections.Service) *Service {
	svc := NewService(d.Queries, inspSvc)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/violations")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.GET("/:id", middleware.RequirePermission(constants.PermInspectionsRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		v, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, v)
	})

	g.PATCH("/:id", middleware.RequirePermission(constants.PermViolationsWrite), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req UpdateViolationRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		user, _ := middleware.CurrentUser(c)
		violation, err := svc.Update(c.Request.Context(), id, req, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionREVIEW, EntityType: "Violation", EntityID: id.String(), NewState: violation})
		response.OK(c, violation)
	})

	return svc
}
