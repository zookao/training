package admin

import (
	adminApi "training/api/admin"

	"github.com/gin-gonic/gin"
)

// RegisterUploadRoutes 分片上传路由（登录即可，免接口权限：视频/课件等共用）
func RegisterUploadRoutes(r *gin.RouterGroup) {
	r.POST("/upload/chunk/init", adminApi.UploadChunkInit)
	r.POST("/upload/chunk", adminApi.UploadChunk)
	r.POST("/upload/chunk/merge", adminApi.UploadChunkMerge)
}

// RegisterImageUploadRoute 图片上传路由（登录即可，免接口权限：课程封面/班级封面等共用）
func RegisterImageUploadRoute(r *gin.RouterGroup) {
	r.POST("/upload/image", adminApi.UploadImage)
}
