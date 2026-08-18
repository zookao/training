package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterCourseRoutes 课程路由
func RegisterCourseRoutes(r *gin.RouterGroup) {
	r.GET("/course", adminApi.CourseList)
	r.GET("/course/all", adminApi.CourseAll)
	r.POST("/course", adminApi.CourseCreate)
	r.GET("/course/:id", adminApi.CourseDetail)
	r.PUT("/course/:id", adminApi.CourseUpdate)
	r.DELETE("/course/:id", adminApi.CourseDelete)
	r.DELETE("/course/:id/video/:vid", adminApi.VideoDelete)
	r.POST("/course/video/:vid/reparse", adminApi.ReparseCoursewarePages)
}

// RegisterDurationRoutes 课件时长导入路由（登录即可，免接口权限）
func RegisterDurationRoutes(r *gin.RouterGroup) {
	r.GET("/course/video/duration-template", adminApi.DurationTemplate)
	r.POST("/course/video/import-durations", adminApi.ImportDurations)
}
