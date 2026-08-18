package admin

import (
	adminModel "training/model/admin"
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// RoleList 角色列表
func RoleList(c *gin.Context) {
	var req common.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	guard := c.DefaultQuery("guardType", "admin")
	res, err := adminService.RolePage(req, guard)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// RoleAll 全部角色（下拉）
func RoleAll(c *gin.Context) {
	guard := c.DefaultQuery("guardType", "admin")
	list, err := adminService.RoleAll(guard)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// RoleDetail 角色详情
func RoleDetail(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	r, err := adminService.RoleGet(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, r)
}

// RoleMenuIDs 角色已分配菜单ID
func RoleMenuIDs(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	ids, err := adminService.RoleGetMenuIDs(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, ids)
}

// RoleApiIDs 角色已分配接口ID
func RoleApiIDs(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	ids, err := adminService.RoleGetApiIDs(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, ids)
}

// RoleCreate 创建角色
func RoleCreate(c *gin.Context) {
	var req adminModel.RoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.RoleCreate(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "创建成功")
}

// RoleUpdate 更新角色
func RoleUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.RoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.RoleUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "更新成功")
}

// RoleDelete 删除角色
func RoleDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.RoleDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "删除成功")
}

// RoleAssignMenus 分配角色菜单
func RoleAssignMenus(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.AssignIDs
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.AssignRoleMenus(id, req.IDs); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "分配成功")
}

// RoleAssignApis 分配角色接口
func RoleAssignApis(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.AssignIDs
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.AssignRoleApis(id, req.IDs); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "分配成功")
}
