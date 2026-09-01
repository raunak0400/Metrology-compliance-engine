// Package response provides the standard API response envelope:
//
//	{ "success": true, "data": {} }
//	{ "success": false, "error": { "code", "message", "details", "requestId" } }
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Success struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    interface{} `json:"meta,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Success{Success: true, Data: data})
}

func OKWithMeta(c *gin.Context, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, Success{Success: true, Data: data, Meta: meta})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Success{Success: true, Data: data})
}

func Accepted(c *gin.Context, data interface{}) {
	c.JSON(http.StatusAccepted, Success{Success: true, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
