package initialize

import (
	"training/middleware"
	adminRouter "training/router/admin"
	userRouter "training/router/user"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册全部路由
func RegisterRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 静态资源：视频/缩略图文件访问（需 JWT 认证，支持 ?token=xxx 查询参数）
	r.Group("/upload", middleware.JWTAuth()).Static("/", "upload")

	// 管理后台 API
	admin := r.Group("/api/admin")

	// 1. 公开路由（登录）
	adminRouter.RegisterAuthPublic(admin)

	// 2. 已登录路由（无需接口鉴权：userinfo/menus/logout/dashboard/导入模板）
	authed := admin.Group("", middleware.JWTAuth(), middleware.GuardType("admin"))
	adminRouter.RegisterAuthAuthed(authed)
	adminRouter.RegisterDashboardRoutes(authed)
	adminRouter.RegisterUserTemplateRoute(authed)
	adminRouter.RegisterDepartmentTemplateRoute(authed)
	adminRouter.RegisterQuestionTemplateRoute(authed)
	adminRouter.RegisterImageUploadRoute(authed)
	adminRouter.RegisterUploadRoutes(authed)
	adminRouter.RegisterDurationRoutes(authed)
	adminRouter.RegisterSettingAuthedRoutes(authed)

	// 3. 已登录 + 接口权限校验
	perm := admin.Group("", middleware.JWTAuth(), middleware.GuardType("admin"), middleware.HasApiPerm())
	adminRouter.RegisterAdminRoutes(perm)
	adminRouter.RegisterRoleRoutes(perm)
	adminRouter.RegisterMenuRoutes(perm)
	adminRouter.RegisterApiRoutes(perm)
	adminRouter.RegisterCourseRoutes(perm)
	adminRouter.RegisterUserRoutes(perm)
	adminRouter.RegisterClassRoutes(perm)
	adminRouter.RegisterDepartmentRoutes(perm)
	adminRouter.RegisterQuestionRoutes(perm)
	adminRouter.RegisterTestpaperRoutes(perm)
	adminRouter.RegisterExamReportRoutes(perm)
	adminRouter.RegisterSettingPermRoutes(perm)

	// 学员前台 API
	user := r.Group("/api/user")
	// 1. 公开路由（注册/登录/院系下拉）
	userRouter.RegisterAuthPublic(user)
	userRouter.RegisterDepartmentRoutes(user)
	userRouter.RegisterSiteRoutes(user)
	// 2. 已登录路由（资料/密码/登出/学习）
	userAuthed := user.Group("", middleware.JWTAuth(), middleware.GuardType("user"))
	userRouter.RegisterAuthAuthed(userAuthed)
	userRouter.RegisterLearningRoutes(userAuthed)
	userRouter.RegisterExamRoutes(userAuthed)
}

