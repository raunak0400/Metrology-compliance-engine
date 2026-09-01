package products

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
	svc := NewService(d.Queries)
	auditSvc := audit.NewService(d.Queries)

	g := rg.Group("/products")
	g.Use(middleware.Authenticate(d.Tokenizer))

	g.POST("", middleware.RequirePermission(constants.PermProductsWrite), func(c *gin.Context) {
		var req CreateProductRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		user, _ := middleware.CurrentUser(c)
		product, err := svc.Create(c.Request.Context(), req, user.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionCREATE, EntityType: "Product", EntityID: product.ID.String(), NewState: product})
		response.Created(c, product)
	})

	g.GET("", middleware.RequirePermission(constants.PermProductsRead), func(c *gin.Context) {
		var q ListProductsQuery
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

	g.GET("/:id", middleware.RequirePermission(constants.PermProductsRead), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		product, err := svc.Get(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, product)
	})

	g.PATCH("/:id", middleware.RequirePermission(constants.PermProductsWrite), func(c *gin.Context) {
		id, err := httpapi.ParseUUIDParam(c, "id")
		if err != nil {
			_ = c.Error(err)
			return
		}
		var req UpdateProductRequest
		if err := httpapi.BindJSON(c, &req); err != nil {
			_ = c.Error(err)
			return
		}
		product, err := svc.Update(c.Request.Context(), id, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		auditSvc.Record(c, audit.RecordParams{Action: db.AuditActionUPDATE, EntityType: "Product", EntityID: id.String(), NewState: product})
		response.OK(c, product)
	})
}
