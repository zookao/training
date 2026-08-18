package admin

import (
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// Dashboard 首页统计
func Dashboard(c *gin.Context) {
	res, err := adminService.Dashboard()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}
