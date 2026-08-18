package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterRoleRoutes 角色路由
func RegisterRoleRoutes(r *gin.RouterGroup) {
	r.GET("/role", adminApi.RoleList)
	r.GET("/role/all", adminApi.RoleAll)
	r.GET("/role/:id", adminApi.RoleDetail)
	r.GET("/role/:id/menuIds", adminApi.RoleMenuIDs)
	r.GET("/role/:id/apiIds", adminApi.RoleApiIDs)
	r.POST("/role", adminApi.RoleCreate)
	r.PUT("/role/:id", adminApi.RoleUpdate)
	r.DELETE("/role/:id", adminApi.RoleDelete)
	r.PUT("/role/:id/menus", adminApi.RoleAssignMenus)
	r.PUT("/role/:id/apis", adminApi.RoleAssignApis)
}
