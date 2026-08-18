package user

import (
	"training/middleware"
	"training/model/common"
	userModel "training/model/user"
	userService "training/service/user"

	"github.com/gin-gonic/gin"
)

// MyClasses 我的班级
func MyClasses(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	list, err := userService.MyClasses(uid)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// ClassDetail 班级详情（含课程进度）
func ClassDetail(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := userService.ClassDetail(uid, id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// CourseLearn 课程学习详情
func CourseLearn(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := userService.CourseLearn(uid, id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// ReportProgress 上报学习进度
func ReportProgress(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	var req userModel.ProgressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := userService.ReportProgress(uid, req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// CheckPass 通过学习滑动校验
func CheckPass(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	videoID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	nextPos, err := userService.CheckPass(uid, videoID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, gin.H{"nextCheckPosition": nextPos})
}
