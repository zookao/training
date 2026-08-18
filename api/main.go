package main

import (
	"log"

	"training/config"
	"training/core"
	"training/global"
	"training/initialize"
	userService "training/service/user"
	"training/utils"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("[Config] 加载失败: %v", err)
	}
	global.Config = cfg

	// 2. 检测 LibreOffice（课件转 PDF 的硬性依赖，未安装则拒绝启动）
	if soffice, err := utils.FindSoffice(); err != nil {
		log.Fatalf("[LibreOffice] %v", err)
	} else {
		log.Printf("[LibreOffice] 检测成功: %s", soffice)
	}

	// 3. 检测 FFmpeg（视频时长获取与封面截取的硬性依赖，未安装则拒绝启动）
	if ffmpeg, err := utils.FindFFmpeg(); err != nil {
		log.Fatalf("[FFmpeg] %v", err)
	} else {
		log.Printf("[FFmpeg] 检测成功: %s", ffmpeg)
	}

	// 4. 连接数据库（首次启动自动创建）
	global.DB = core.Mysql()

	// 5. 自动迁移 + 种子
	initialize.Migrate(global.DB)
	initialize.Seed(global.DB)

	// 6. 启动考试超时清理定时任务（后台协程，依赖 DB 已就绪）
	userService.StartExamScheduler()

	// 7. 启动服务
	core.RunServer()
}
