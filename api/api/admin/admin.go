package admin

import (
	adminModel "training/model/admin"
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// AdminList 管理员列表
func AdminList(c *gin.Context) {
	var req common.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := adminService.AdminPage(req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// AdminCreate 创建管理员
func AdminCreate(c *gin.Context) {
	var req adminModel.AdminCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.AdminCreate(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "创建成功")
}

// AdminDetail 管理员详情
func AdminDetail(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	a, err := adminService.AdminGet(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, a)
}

// AdminUpdate 更新管理员
func AdminUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.AdminUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.AdminUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "更新成功")
}

// AdminRoles 分配管理员角色
func AdminRoles(c *gin.Context) {
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
	if err := adminService.AssignRoles(id, req.IDs); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "分配成功")
}

// AdminResetPwd 重置密码
func AdminResetPwd(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.AdminResetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.AdminResetPassword(id, req.Password); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "重置成功")
}

// AdminDelete 删除管理员
func AdminDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.AdminDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "删除成功")
}
