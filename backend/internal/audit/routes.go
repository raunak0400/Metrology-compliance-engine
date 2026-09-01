package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/pagination"
	"github.com/raunak0400/metrology-backend/internal/response"
)

func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps) *Service {
	svc := NewService(d.Queries)

	g := rg.Group("/audit-logs")
	g.Use(middleware.Authenticate(d.Tokenizer), middleware.RequirePermission(constants.PermAuditRead))

	g.GET("", func(c *gin.Context) {
		page := pagination.Parse(c)
		q := ListQuery{
			EntityType: c.Query("entityType"),
			EntityID:   c.Query("entityId"),
			ActorID:    c.Query("actorId"),
		}
		items, total, err := svc.List(c.Request.Context(), q, page)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OKWithMeta(c, items, pagination.Build(items, total, page))
	})

	return svc
}
