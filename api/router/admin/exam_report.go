package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterExamReportRoutes 注册考试报告路由
func RegisterExamReportRoutes(rg *gin.RouterGroup) {
	rg.GET("/course/:id/exam-report", adminApi.ExamReport)
	rg.GET("/exam-record/:id", adminApi.ExamRecordDetail)
}
