package admin

import (
	"log"
	"time"

	"training/utils"
)

// 上传残留清扫器：周期清理 .chunks 下管理员的过期分片会话
// （上传中断/放弃后永久残留的临时分片，不属于已保存的业务内容）
//
// 注意：不清理 upload/ 下的孤儿文件。交付要求"删除系统内容不删除上传文件"，
// 已上传落盘的文件（视频/课件/封面等）一律保留，由磁盘容量规划与人工治理。
const (
	chunkStaleTTLSweep = 24 * time.Hour  // 分片会话过期阈值（与上传侧 chunkStaleTTL 一致）
	sweepInterval      = 1 * time.Hour   // 清扫周期
	firstSweepDelay    = 10 * time.Minute // 启动后延迟首轮，避免与启动初始化抢 IO
)

// StartUploadSweeper 启动上传残留清扫定时任务（后台协程，依赖 DB 已就绪）
func StartUploadSweeper() {
	go func() {
		time.Sleep(firstSweepDelay)
		SweepUploadResidue()
		ticker := time.NewTicker(sweepInterval)
		defer ticker.Stop()
		log.Printf("[UploadSweeper] 已启动，每 %s 清扫一次过期分片会话", sweepInterval)
		for range ticker.C {
			SweepUploadResidue()
		}
	}()
}

// SweepUploadResidue 执行一轮清扫（导出以便手动触发/测试）
func SweepUploadResidue() {
	if n := utils.SweepStaleChunks(chunkStaleTTLSweep); n > 0 {
		log.Printf("[UploadSweeper] 清理过期分片会话 %d 个", n)
	}
}
