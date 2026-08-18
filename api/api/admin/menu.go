package admin

import (
	adminModel "training/model/admin"
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// MenuList 菜单树
func MenuList(c *gin.Context) {
	guard := c.DefaultQuery("guardType", "admin")
	menus, err := adminService.MenuList(guard)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, menus)
}

// MenuDetail 菜单详情
func MenuDetail(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	m, err := adminService.MenuGet(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, m)
}

// MenuCreate 创建菜单
func MenuCreate(c *gin.Context) {
	var req adminModel.MenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.MenuCreate(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "创建成功")
}

// MenuUpdate 更新菜单
func MenuUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.MenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.MenuUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "更新成功")
}

// MenuDelete 删除菜单
func MenuDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.MenuDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "删除成功")
}
