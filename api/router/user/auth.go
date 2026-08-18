package user

import (
	userApi "training/api/user"

	"github.com/gin-gonic/gin"
)

// RegisterAuthPublic 学员公开路由（注册/登录）
func RegisterAuthPublic(r *gin.RouterGroup) {
	r.POST("/auth/register", userApi.Register)
	r.POST("/auth/login", userApi.Login)
}

// RegisterAuthAuthed 学员已登录路由（资料/密码/登出）
func RegisterAuthAuthed(r *gin.RouterGroup) {
	r.GET("/auth/userinfo", userApi.UserInfo)
	r.PUT("/auth/profile", userApi.UpdateProfile)
	r.PUT("/auth/password", userApi.ChangePassword)
	r.POST("/auth/logout", userApi.Logout)
}
