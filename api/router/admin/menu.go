package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterMenuRoutes 菜单路由
func RegisterMenuRoutes(r *gin.RouterGroup) {
	r.GET("/menu", adminApi.MenuList)
	r.GET("/menu/:id", adminApi.MenuDetail)
	r.POST("/menu", adminApi.MenuCreate)
	r.PUT("/menu/:id", adminApi.MenuUpdate)
	r.DELETE("/menu/:id", adminApi.MenuDelete)
}
