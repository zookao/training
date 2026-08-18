package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterTestpaperRoutes 注册试卷路由
func RegisterTestpaperRoutes(rg *gin.RouterGroup) {
	rg.GET("/course/:id/testpaper", adminApi.TestpaperList)
	rg.POST("/course/:id/testpaper", adminApi.TestpaperCreate)
	rg.PUT("/testpaper/:id", adminApi.TestpaperUpdate)
	rg.DELETE("/testpaper/:id", adminApi.TestpaperDelete)
	rg.GET("/testpaper/:id/questions", adminApi.TestpaperGetQuestions)
	rg.PUT("/testpaper/:id/questions", adminApi.TestpaperSetQuestions)
}
