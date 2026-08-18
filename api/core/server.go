package core

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"training/global"
	"training/initialize"

	"github.com/gin-gonic/gin"
)

// RunServer 启动 HTTP 服务
func RunServer() {
	if global.Config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	// 大视频上传：超出 8MB 走临时文件，避免占用过多内存
	r.MaxMultipartMemory = 8 << 20

	initialize.RegisterRoutes(r)

	port := global.Config.Server.Port
	log.Printf("[Server] 监听 :%d", port)
	// 显式 http.Server：
	//   - ReadHeaderTimeout 10s：防止 slowloris 慢速头部攻击
	//   - ReadTimeout/WriteTimeout 为 0（不限时）：支持大文件上传（>1GB），上传耗时由 nginx 层控制
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[Server] 启动失败: %v", err)
	}
}
