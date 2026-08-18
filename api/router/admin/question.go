package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterQuestionRoutes 注册试题路由
func RegisterQuestionRoutes(rg *gin.RouterGroup) {
	rg.GET("/course/:id/question", adminApi.QuestionList)
	rg.GET("/course/:id/question/all", adminApi.QuestionAll)
	rg.POST("/course/:id/question", adminApi.QuestionCreate)
	rg.POST("/course/:id/question/import", adminApi.QuestionImport)
	rg.PUT("/question/:id", adminApi.QuestionUpdate)
	rg.DELETE("/question/:id", adminApi.QuestionDelete)
}

// RegisterQuestionTemplateRoute 试题导入模板下载路由（已登录即可访问，无需接口鉴权）
// 注意：挂在 /template/ 下，避免与 /course/:id 等通配路由冲突。
func RegisterQuestionTemplateRoute(r *gin.RouterGroup) {
	r.GET("/template/question-import", adminApi.QuestionImportTemplate)
}
