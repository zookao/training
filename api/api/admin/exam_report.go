package admin

import (
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// ExamReport 考试报告
func ExamReport(c *gin.Context) {
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	keyword := c.Query("keyword")
	list, err := adminService.ExamReport(courseID, keyword)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// ExamRecordDetail 考试记录详情
func ExamRecordDetail(c *gin.Context) {
	recordID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := adminService.ExamRecordDetail(recordID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}
