package inspections

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/audit"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/httpapi"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/pagination"
	"github.com/raunak0400/metrology-backend/internal/response"
)

func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps) *Service {
	svc := NewService(d.Queries)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/inspections")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.POST("", middleware.RequirePermission(constants.PermInspectionsWrite), func(c *gin.Context) {
		var req CreateInspectionRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		user, ok := middleware.CurrentUser(c)
		if !ok {
			_ = c.Error(apperror.Unauthorized(""))
			return
		}
		inspectorID, _ := uuid.Parse(user.ID)
		inspection, err := svc.Create(c.Request.Context(), req, inspectorID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionCREATE, EntityType: "Inspection", EntityID: inspection.ID.String(), NewState: inspection})
		response.Created(c, inspection)
	})

	g.GET("", middleware.RequirePermission(constants.PermInspectionsRead), func(c *gin.Context) {
		var q ListInspectionsQuery
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

	g.GET("/:id", middleware.RequirePermission(constants.PermInspectionsRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		inspection, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, inspection)
	})

	g.PATCH("/:id", middleware.RequirePermission(constants.PermInspectionsWrite), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req UpdateInspectionRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		inspection, err := svc.Update(c.Request.Context(), id, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionUPDATE, EntityType: "Inspection", EntityID: id.String(), NewState: inspection})
		response.OK(c, inspection)
	})

	// Finalize/reopen: a controlled, audited alternative to the generic PATCH
	// above. Once finalized, PATCH (and declaration/violation edits) are
	// rejected until the inspection is explicitly reopened.
	g.POST("/:id/finalize", middleware.RequirePermission(constants.PermInspectionsFinalize), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		user, ok := middleware.CurrentUser(c)
		if !ok {
			_ = c.Error(apperror.Unauthorized(""))
			return
		}
		inspection, err := svc.Finalize(c.Request.Context(), id, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionFINALIZE, EntityType: "Inspection", EntityID: id.String(), NewState: inspection})
		response.OK(c, inspection)
	})

	g.POST("/:id/reopen", middleware.RequirePermission(constants.PermInspectionsFinalize), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		inspection, err := svc.Reopen(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionUPDATE, EntityType: "Inspection", EntityID: id.String(), NewState: inspection})
		response.OK(c, inspection)
	})

	return svc
}
