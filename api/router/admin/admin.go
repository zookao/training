package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes 管理员路由
func RegisterAdminRoutes(r *gin.RouterGroup) {
	r.GET("/admin", adminApi.AdminList)
	r.POST("/admin", adminApi.AdminCreate)
	r.GET("/admin/:id", adminApi.AdminDetail)
	r.PUT("/admin/:id", adminApi.AdminUpdate)
	r.PUT("/admin/:id/roles", adminApi.AdminRoles)
	r.PUT("/admin/:id/password", adminApi.AdminResetPwd)
	r.DELETE("/admin/:id", adminApi.AdminDelete)
}
