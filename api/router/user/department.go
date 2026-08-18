package user

import (
	userApi "training/api/user"

	"github.com/gin-gonic/gin"
)

// RegisterDepartmentRoutes 注册院系公开路由（注册页下拉用，无需登录）
func RegisterDepartmentRoutes(r *gin.RouterGroup) {
	r.GET("/department/all", userApi.DepartmentList)
}
