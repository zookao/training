package user

import (
	"os"

	"training/model/common"
	adminService "training/service/admin"
	"training/utils"

	"github.com/gin-gonic/gin"
)

// GetSiteConfig 公开接口：站点配置（logo + 站点名称，登录页/注册页用）
func GetSiteConfig(c *gin.Context) {
	logoUrl, siteName := adminService.GetSiteConfig()
	common.OK(c, gin.H{"logoUrl": logoUrl, "siteName": siteName})
}

// SiteLogo 公开接口：返回 logo 图片（无需 token，供未登录的登录页/注册页 <img> 直接用）
func SiteLogo(c *gin.Context) {
	logoUrl, _ := adminService.GetSiteConfig()
	if logoUrl == "" {
		c.Status(404)
		return
	}
	localPath, ok := utils.URLToLocalPath(logoUrl)
	if !ok {
		c.Status(404)
		return
	}
	if _, err := os.Stat(localPath); err != nil {
		c.Status(404)
		return
	}
	c.File(localPath)
}
