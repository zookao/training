package admin

import (
	adminModel "training/model/admin"
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// ApiList 接口列表
func ApiList(c *gin.Context) {
	var req common.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := adminService.ApiPage(req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// ApiAll 全部接口
func ApiAll(c *gin.Context) {
	list, err := adminService.ApiAll()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// ApiCreate 创建接口
func ApiCreate(c *gin.Context) {
	var req adminModel.ApiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.ApiCreate(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "创建成功")
}

// ApiUpdate 更新接口
func ApiUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.ApiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.ApiUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "更新成功")
}

// ApiDelete 删除接口
func ApiDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.ApiDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "删除成功")
}
