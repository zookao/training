package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterApiRoutes 接口路由
func RegisterApiRoutes(r *gin.RouterGroup) {
	r.GET("/api", adminApi.ApiList)
	r.GET("/api/all", adminApi.ApiAll)
	r.POST("/api", adminApi.ApiCreate)
	r.PUT("/api/:id", adminApi.ApiUpdate)
	r.DELETE("/api/:id", adminApi.ApiDelete)
}
