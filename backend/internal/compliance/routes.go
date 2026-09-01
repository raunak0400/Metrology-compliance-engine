package compliance

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
	"github.com/raunak0400/metrology-backend/internal/violations"
)

// RegisterRoutes owns /inspections/:id/check, /:id/checks, /:id/violations.
func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps, inspSvc *inspections.Service, violSvc *violations.Service) *Service {
	svc := NewService(d.Queries, d.Asynq, inspSvc)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/inspections")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.POST("/:id/check", middleware.RequirePermission(constants.PermInspectionsWrite), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		user, _ := middleware.CurrentUser(c)
		jobID, err := svc.Enqueue(c.Request.Context(), id, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionPROCESS, EntityType: "Inspection", EntityID: id.String()})
		response.Accepted(c, gin.H{"jobId": jobID, "status": "QUEUED"})
	})

	g.GET("/:id/checks", middleware.RequirePermission(constants.PermInspectionsRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		checks, err := svc.ListChecks(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, checks)
	})

	g.GET("/:id/violations", middleware.RequirePermission(constants.PermInspectionsRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		items, err := violSvc.ListByInspection(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, items)
	})

	return svc
}
