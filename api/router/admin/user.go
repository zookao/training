package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 学员管理路由
func RegisterUserRoutes(r *gin.RouterGroup) {
	r.GET("/user", adminApi.UserList)
	r.GET("/user/all", adminApi.UserAll)
	r.GET("/user/:id", adminApi.UserDetail)
	r.POST("/user", adminApi.UserCreate)
	r.POST("/user/import", adminApi.UserImport)
	r.PUT("/user/:id", adminApi.UserUpdate)
	r.PUT("/user/:id/password", adminApi.UserResetPwd)
	r.DELETE("/user/:id", adminApi.UserDelete)
}

// RegisterUserTemplateRoute 学员导入模板下载路由（已登录即可访问）
// 注意：不能挂在 /user/ 下，否则与 /user/:id 通配路由冲突。
func RegisterUserTemplateRoute(r *gin.RouterGroup) {
	r.GET("/template/user-import", adminApi.UserImportTemplate)
}
