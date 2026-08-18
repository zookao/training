package user

import (
	userApi "training/api/user"

	"github.com/gin-gonic/gin"
)

// RegisterExamRoutes 学员考试路由
func RegisterExamRoutes(r *gin.RouterGroup) {
	r.GET("/course/:id/exam", userApi.GetCourseExam)
	r.GET("/testpaper/:id/exam", userApi.GetExam)
	r.POST("/testpaper/:id/draft", userApi.SaveExamDraft)
	r.POST("/testpaper/:id/submit", userApi.SubmitExam)
	r.GET("/course/:id/exam/records", userApi.GetExamRecords)
}
