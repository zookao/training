package user

import (
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// DepartmentList 公开接口：全部启用院系（注册时选择用）
func DepartmentList(c *gin.Context) {
	list, err := adminService.DepartmentAll()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}
