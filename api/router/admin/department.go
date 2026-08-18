package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterDepartmentRoutes 注册院系路由
func RegisterDepartmentRoutes(rg *gin.RouterGroup) {
	rg.GET("/department", adminApi.DepartmentList)
	rg.GET("/department/all", adminApi.DepartmentAll)
	rg.POST("/department", adminApi.DepartmentCreate)
	rg.PUT("/department/:id", adminApi.DepartmentUpdate)
	rg.DELETE("/department/:id", adminApi.DepartmentDelete)
	// 院系学员管理
	rg.GET("/department/:id/students", adminApi.DepartmentStudentList)
	rg.POST("/department/:id/students/import", adminApi.DepartmentStudentImport)
	rg.DELETE("/department/:id/students/:userId", adminApi.DepartmentStudentRemove)
}

// RegisterDepartmentTemplateRoute 院系学员导入模板下载路由（已登录即可访问）
// 注意：不能挂在 /department/ 下，否则与 /department/:id 通配路由冲突。
func RegisterDepartmentTemplateRoute(r *gin.RouterGroup) {
	r.GET("/template/department-student-import", adminApi.DepartmentStudentImportTemplate)
}
