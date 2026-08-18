package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterDashboardRoutes 首页统计路由（已登录即可访问，无需接口鉴权）
func RegisterDashboardRoutes(r *gin.RouterGroup) {
	r.GET("/dashboard", adminApi.Dashboard)
}
