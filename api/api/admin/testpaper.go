package admin

import (
	"training/model/admin"
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// TestpaperList 试卷列表
func TestpaperList(c *gin.Context) {
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req common.PageRequest
	_ = c.ShouldBindQuery(&req)
	req.Normalize()
	res, err := adminService.TestpaperList(courseID, req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// TestpaperCreate 创建试卷
func TestpaperCreate(c *gin.Context) {
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req admin.TestpaperReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.TestpaperCreate(courseID, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// TestpaperUpdate 更新试卷
func TestpaperUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req admin.TestpaperReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.TestpaperUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// TestpaperDelete 删除试卷
func TestpaperDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.TestpaperDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// TestpaperGetQuestions 获取试卷试题
func TestpaperGetQuestions(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	list, err := adminService.TestpaperGetQuestions(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// TestpaperSetQuestions 设置试卷试题
func TestpaperSetQuestions(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req admin.TestpaperQuestionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.TestpaperSetQuestions(id, req.Items); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}
