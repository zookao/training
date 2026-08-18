package user

import (
	userApi "training/api/user"

	"github.com/gin-gonic/gin"
)

// RegisterSiteRoutes 站点配置公开路由（登录页/注册页用，无需登录）
func RegisterSiteRoutes(r *gin.RouterGroup) {
	r.GET("/site-config", userApi.GetSiteConfig)
	r.GET("/site-logo", userApi.SiteLogo)
}
