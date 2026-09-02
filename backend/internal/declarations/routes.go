package declarations

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/audit"
	"github.com/raunak0400/metrology-backend/internal/compliance"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/httpapi"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/response"
)

// RegisterRoutes owns /inspections/:id/declarations and /declarations/:id.
// complianceSvc is used to re-enqueue a compliance evaluation whenever a
// reviewer edits a declaration — the prior compliance checks/violations were
// computed against the old value and are now stale.
func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps, inspSvc *inspections.Service, complianceSvc *compliance.Service) *Service {
	svc := NewService(d.Queries, inspSvc)
	auditSvc := audit.NewService(d.Queries)

	authGroup := rg.Group("/")
	authGroup.Use(middleware.Authenticate(d.Tokenizer))

	authGroup.GET("/inspections/:id/declarations", middleware.RequirePermission(constants.PermInspectionsRead), func(c *gin.Context) {
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

	authGroup.PATCH("/declarations/:id", middleware.RequirePermission(constants.PermDeclarationsWrite), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req UpdateDeclarationRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		before, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		user, _ := middleware.CurrentUser(c)
		declaration, err := svc.Update(c.Request.Context(), id, req, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionREVIEW, EntityType: "ExtractedDeclaration", EntityID: id.String(), OldState: before, NewState: declaration})

		// The prior compliance checks/violations were computed against the old
		// declaration value and are now stale — re-run the deterministic engine.
		if _, err := complianceSvc.Enqueue(c.Request.Context(), declaration.InspectionID, user.ID); err != nil {
			slog.Error("failed to enqueue re-evaluation after declaration edit", "err", err, "declarationId", id, "inspectionId", declaration.InspectionID)
		}

		response.OK(c, declaration)
	})

	return svc
}
