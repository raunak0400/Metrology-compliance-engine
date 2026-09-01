package evidence

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

// RegisterRoutes owns /inspections/:id/evidence.
func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps, inspSvc *inspections.Service) {
	svc := NewService(d.Queries, d.Storage, inspSvc)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/inspections")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.POST(
		"/:id/evidence",
		middleware.RequirePermission(constants.PermInspectionsWrite),
		// Optional file upload on field `file`; JSON-only references are also allowed.
		middleware.RequireSingleImage("file", d.Config.MaxUploadSizeBytes, d.Config.AllowedImageMIME),
		func(c *gin.Context) {
			id, err := httpapi.ParseUUIDParam(c, "id")
			if err != nil {
				_ = c.Error(err)
				return
			}
			var req AddEvidenceRequest
			if err := httpapi.Bind(c, &req); err != nil {
				_ = c.Error(err)
				return
			}
			user, _ := middleware.CurrentUser(c)

			var buf []byte
			var filename, contentType string
			if file, ok := middleware.UploadedFileFromContext(c); ok {
				buf, filename, contentType = file.Buffer, file.Filename, file.ContentType
			}

			ev, err := svc.Add(c.Request.Context(), id, req, user.ID, buf, filename, contentType)
			if err != nil {
				_ = c.Error(err)
				return
			}
			auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionCREATE, EntityType: "Evidence", EntityID: ev.ID.String()})
			response.Created(c, ev)
		},
	)

	g.GET("/:id/evidence", middleware.RequirePermission(constants.PermInspectionsRead), func(c *gin.Context) {
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
}
