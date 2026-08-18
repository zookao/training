package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterAuthPublic 公开路由（登录）
func RegisterAuthPublic(r *gin.RouterGroup) {
	r.POST("/auth/login", adminApi.Login)
}

// RegisterAuthAuthed 已登录路由（userinfo/menus/logout，无接口鉴权）
func RegisterAuthAuthed(r *gin.RouterGroup) {
	r.GET("/auth/userinfo", adminApi.UserInfo)
	r.GET("/auth/menus", adminApi.Menus)
	r.POST("/auth/logout", adminApi.Logout)
}
