package reports

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

// RegisterRoutes owns /inspections/:id/reports and /reports/:id.
func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps, inspSvc *inspections.Service) *Service {
	svc := NewService(d.Queries, d.Storage, d.Asynq, inspSvc)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.POST("/inspections/:id/reports", middleware.RequirePermission(constants.PermReportsRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req CreateReportRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		user, _ := middleware.CurrentUser(c)
		result, err := svc.Request(c.Request.Context(), id, req.Type, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionEXPORT, EntityType: "Report", EntityID: result.ReportID})
		response.Created(c, result)
	})

	g.GET("/inspections/:id/reports", middleware.RequirePermission(constants.PermReportsRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		items, err := svc.ListByInspection(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, items)
	})

	g.GET("/reports/:id", middleware.RequirePermission(constants.PermReportsRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		report, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, report)
	})

	return svc
}
