package datasets

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

// RegisterRoutes owns /datasets, /datasets/:id/items and
// /dataset-items/:id/annotations (mounted at the API root).
func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps) {
	svc := NewService(d.Queries)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.POST("/datasets", middleware.RequirePermission(constants.PermDatasetsWrite), func(c *gin.Context) {
		var req CreateDatasetRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		user, _ := middleware.CurrentUser(c)
		dataset, err := svc.Create(c.Request.Context(), req, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionCREATE, EntityType: "Dataset", EntityID: dataset.ID.String()})
		response.Created(c, dataset)
	})

	g.GET("/datasets", func(c *gin.Context) {
		page := pagination.Parse(c)
		items, total, err := svc.List(c.Request.Context(), page)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OKWithMeta(c, items, pagination.Build(items, total, page))
	})

	g.POST("/datasets/:id/items", middleware.RequirePermission(constants.PermDatasetsWrite), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req AddItemRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		item, err := svc.AddItem(c.Request.Context(), id, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.Created(c, item)
	})

	g.GET("/datasets/:id/items", func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var q ListQuery
		if err := httpapi.BindQuery(c, &q); err != nil {
			_ = c.Error(err)
			return
		}
		page := pagination.Parse(c)
		items, total, err := svc.ListItems(c.Request.Context(), id, q, page)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OKWithMeta(c, items, pagination.Build(items, total, page))
	})

	g.POST("/dataset-items/:id/annotations", middleware.RequirePermission(constants.PermDatasetsWrite), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req AddAnnotationRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		user, _ := middleware.CurrentUser(c)
		annotation, err := svc.AddAnnotation(c.Request.Context(), id, req, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.Created(c, annotation)
	})
}
