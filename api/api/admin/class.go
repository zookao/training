package admin

import (
	adminModel "training/model/admin"
	"training/model/common"
	adminService "training/service/admin"
	userLearning "training/service/user"

	"github.com/gin-gonic/gin"
)

// ClassList 班级列表
func ClassList(c *gin.Context) {
	var req common.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := adminService.ClassPage(req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// ClassDetail 班级详情
func ClassDetail(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	cl, err := adminService.ClassGet(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, cl)
}

// ClassCreate 创建班级
func ClassCreate(c *gin.Context) {
	var req adminModel.ClassReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.ClassCreate(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "创建成功")
}

// ClassUpdate 更新班级
func ClassUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.ClassReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.ClassUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "更新成功")
}

// ClassDelete 删除班级
func ClassDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.ClassDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "删除成功")
}

// ClassCourseIDsHandler 班级已绑课程ID
func ClassCourseIDsHandler(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	ids, err := adminService.ClassCourseIDs(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, ids)
}

// ClassUserIDsHandler 班级已绑学员ID
func ClassUserIDsHandler(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	ids, err := adminService.ClassUserIDs(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, ids)
}

// ClassAssignCourses 分配课程
func ClassAssignCourses(c *gin.Context) {
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
	if err := adminService.AssignClassCourses(id, req.IDs); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "分配成功")
}

// ClassAssignUsers 分配学员
func ClassAssignUsers(c *gin.Context) {
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
	if err := adminService.AssignClassUsers(id, req.IDs); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "分配成功")
}

// ClassLearningReport 班级学习报告
func ClassLearningReport(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := userLearning.ClassLearningReport(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// StudentLearningDetail 学员学习详情
func StudentLearningDetail(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	userID, err := common.ParseIDParam(c, "userId")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := userLearning.StudentLearningDetail(id, userID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}
