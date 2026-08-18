package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterClassRoutes 班级路由
func RegisterClassRoutes(r *gin.RouterGroup) {
	r.GET("/class", adminApi.ClassList)
	r.POST("/class", adminApi.ClassCreate)
	r.GET("/class/:id", adminApi.ClassDetail)
	r.PUT("/class/:id", adminApi.ClassUpdate)
	r.DELETE("/class/:id", adminApi.ClassDelete)
	r.GET("/class/:id/courseIds", adminApi.ClassCourseIDsHandler)
	r.GET("/class/:id/userIds", adminApi.ClassUserIDsHandler)
	r.PUT("/class/:id/courses", adminApi.ClassAssignCourses)
	r.PUT("/class/:id/users", adminApi.ClassAssignUsers)
	r.GET("/class/:id/learning-report", adminApi.ClassLearningReport)
	r.GET("/class/:id/learning-report/:userId", adminApi.StudentLearningDetail)
}
