package middleware

import (
	"training/global"
	"training/model/admin"
	"training/model/common"

	"github.com/gin-gonic/gin"
)

// HasApiPerm 校验当前管理员是否拥有访问该接口的权限
// 通过 admin→roles→apis 关系匹配 method+path
func HasApiPerm() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			common.FailWithCode(c, common.CodeNoAuth, "未登录")
			c.Abort()
			return
		}
		// 超管（id=1）直接放行
		if claims.UserID == 1 {
			c.Next()
			return
		}
		method := c.Request.Method
		path := c.FullPath() // 使用路由模板，如 /api/admin/admin/:id
		if path == "" {
			path = c.Request.URL.Path
		}

		var count int64
		global.DB.Model(&admin.Api{}).
			Joins("JOIN role_apis ON role_apis.api_id = apis.id").
			Joins("JOIN admin_roles ON admin_roles.role_id = role_apis.role_id").
			Where("admin_roles.admin_id = ? AND apis.method = ? AND apis.path = ?", claims.UserID, method, path).
			Count(&count)

		if count == 0 {
			common.FailWithCode(c, common.CodeNoPerm, "无接口访问权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
