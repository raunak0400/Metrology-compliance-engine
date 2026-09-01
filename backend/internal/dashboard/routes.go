package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/response"
)

func RegisterRoutes(rg *gin.RouterGroup, d *app.Deps) {
	svc := NewService(d.Queries)

	g := rg.Group("/dashboard")
	g.Use(middleware.Authenticate(d.Tokenizer), middleware.RequirePermission(constants.PermDashboardRead))

	g.GET("/summary", func(c *gin.Context) {
		summary, err := svc.GetSummary(c.Request.Context())
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, summary)
	})

	g.GET("/violations", func(c *gin.Context) {
		stats, err := svc.GetViolationStats(c.Request.Context())
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, stats)
	})

	g.GET("/inspections", func(c *gin.Context) {
		stats, err := svc.GetInspectionStats(c.Request.Context())
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.OK(c, stats)
	})
}
