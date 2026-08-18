package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterSettingAuthedRoutes 系统设置查询路由（登录即可，免接口权限）
func RegisterSettingAuthedRoutes(r *gin.RouterGroup) {
	r.GET("/setting/site", adminApi.GetSiteConfig)
}

// RegisterSettingPermRoutes 系统设置修改路由（需接口权限）
func RegisterSettingPermRoutes(r *gin.RouterGroup) {
	r.PUT("/setting/site", adminApi.UpdateSiteConfig)
}
