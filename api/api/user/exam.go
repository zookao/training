package user

import (
	"training/middleware"
	"training/model/common"
	userService "training/service/user"

	"github.com/gin-gonic/gin"
)

// GetCourseExam 获取课程下的试卷列表（含可用状态）
func GetCourseExam(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	list, err := userService.GetCourseTestpapers(uid, courseID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// GetExam 获取试卷试题（学员端，不含答案）
func GetExam(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	testpaperID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := userService.GetExam(uid, testpaperID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// SaveExamDraft 保存考试草稿（断点续考）
func SaveExamDraft(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	testpaperID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req userService.SubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := userService.SaveExamDraft(uid, testpaperID, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// SubmitExam 提交考试
func SubmitExam(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	testpaperID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req userService.SubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := userService.SubmitExam(uid, testpaperID, req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// GetExamRecords 获取学员考试记录
func GetExamRecords(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	list, err := userService.GetExamRecords(uid, courseID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}
