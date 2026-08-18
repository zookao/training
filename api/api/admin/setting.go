package admin

import (
	adminService "training/service/admin"
	"training/model/common"

	"github.com/gin-gonic/gin"
)

// GetSiteConfig GET /api/admin/setting/site
func GetSiteConfig(c *gin.Context) {
	logoUrl, siteName := adminService.GetSiteConfig()
	common.OK(c, gin.H{"logoUrl": logoUrl, "siteName": siteName})
}

// UpdateSiteConfig PUT /api/admin/setting/site
func UpdateSiteConfig(c *gin.Context) {
	var req struct {
		LogoUrl  string `json:"logoUrl"`
		SiteName string `json:"siteName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误: "+err.Error())
		return
	}
	if err := adminService.UpdateSiteConfig(req.LogoUrl, req.SiteName); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "保存成功")
}
