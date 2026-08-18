package user

import (
	userApi "training/api/user"

	"github.com/gin-gonic/gin"
)

// RegisterLearningRoutes 学员学习路由
func RegisterLearningRoutes(r *gin.RouterGroup) {
	r.GET("/classes", userApi.MyClasses)
	r.GET("/classes/:id", userApi.ClassDetail)
	r.GET("/course/:id", userApi.CourseLearn)
	r.POST("/progress", userApi.ReportProgress)
	r.POST("/video/:id/check", userApi.CheckPass)
}
