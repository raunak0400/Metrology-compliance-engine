package scans

import (
	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/audit"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/httpapi"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/response"
)

// RegisterRoutes mounts the scan sub-resources under /inspections/:id/...
func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps, inspSvc *inspections.Service) {
	svc := NewService(d.Queries, d.Storage, d.Asynq, inspSvc)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/inspections")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.POST(
		"/:id/images",
		middleware.RequirePermission(constants.PermScansWrite),
		middleware.RequireSingleImage("image", d.Config.MaxUploadSizeBytes, d.Config.AllowedImageMIME),
		func(c *gin.Context) {
			id, err := httpapi.ParseUUIDParam(c, "id")
			if err != nil {
				_ = c.Error(err)
				return
			}
			file, ok := middleware.UploadedFileFromContext(c)
			if !ok {
				_ = c.Error(apperror.BadRequest(`No image file provided (field name: "image")`))
				return
			}
			result, err := svc.AddImage(c.Request.Context(), id, file.Buffer, file.Filename, file.ContentType)
			if err != nil {
				_ = c.Error(err)
				return
			}
			auditSvc.Record(c, audit.RecordParams{
				Action: db.AuditActionCREATE, EntityType: "InspectionImage", EntityID: result.Image.ID.String(),
				NewState: gin.H{"storageKey": result.Image.StorageKey},
			})
			response.Created(c, gin.H{
				"imageId": result.Image.ID, "storageKey": result.Image.StorageKey,
				"jobId": result.JobID, "status": "QUEUED",
			})
		},
	)

	g.POST(
		"/:id/process",
		middleware.RequirePermission(constants.PermScansWrite),
		func(c *gin.Context) {
			id, err := httpapi.ParseUUIDParam(c, "id")
			if err != nil {
				_ = c.Error(err)
				return
			}
			queued, err := svc.ProcessInspection(c.Request.Context(), id)
			if err != nil {
				_ = c.Error(err)
				return
			}
			auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionPROCESS, EntityType: "Inspection", EntityID: id.String()})
			response.Accepted(c, gin.H{"queued": queued})
		},
	)

	g.GET(
		"/:id/jobs",
		middleware.RequirePermission(constants.PermInspectionsRead),
		func(c *gin.Context) {
			id, err := httpapi.ParseUUIDParam(c, "id")
			if err != nil {
				_ = c.Error(err)
				return
			}
			jobs, err := svc.ListJobs(c.Request.Context(), id)
			if err != nil {
				_ = c.Error(err)
				return
			}
			response.OK(c, jobs)
		},
	)
}
