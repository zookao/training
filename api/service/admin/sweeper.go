package admin

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"training/global"
	adminModel "training/model/admin"
	userModel "training/model/user"
	"training/utils"
)

// 上传文件清扫器：周期清理两类磁盘残留
//  1. .chunks 下所有管理员的过期分片会话（上传中断/放弃后永久残留）
//  2. upload/ 下超过宽限期且无任何数据库引用的孤儿文件
//     （分片合并成功但管理员未保存课程、记录删除后因共享引用被跳过等场景）
//
// 孤儿判定是"文件 mtime 超过宽限期 && 全库无引用"双条件：
// 刚合并落盘、课程还没保存的文件处于无引用状态，必须靠宽限期保护。
const (
	chunkStaleTTLSweep = 24 * time.Hour  // 分片会话过期阈值（与上传侧 chunkStaleTTL 一致）
	orphanGrace        = 24 * time.Hour  // 孤儿文件宽限期：落盘超过该时长仍无引用才清理
	sweepInterval      = 1 * time.Hour   // 清扫周期
	firstSweepDelay    = 10 * time.Minute // 启动后延迟首轮，避免与启动初始化抢 IO
)

// uploadURLRe 提取配置值中内嵌的 /upload/... URL（settings.value 可能存 JSON）
var uploadURLRe = regexp.MustCompile(`/upload/[A-Za-z0-9._/-]+`)

// StartUploadSweeper 启动上传残留清扫定时任务（后台协程，依赖 DB 已就绪）
func StartUploadSweeper() {
	go func() {
		time.Sleep(firstSweepDelay)
		SweepUploadResidue()
		ticker := time.NewTicker(sweepInterval)
		defer ticker.Stop()
		log.Printf("[UploadSweeper] 已启动，每 %s 清扫一次分片残留与孤儿文件", sweepInterval)
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
	if n := sweepOrphanUploads(); n > 0 {
		log.Printf("[UploadSweeper] 清理孤儿文件 %d 个", n)
	}
}

// sweepOrphanUploads 删除无数据库引用且超过宽限期的上传文件，返回删除数
func sweepOrphanUploads() int {
	refs, err := collectUploadRefs()
	if err != nil {
		// 引用清单拉取失败时宁可不清：漏判引用会误删线上文件
		log.Printf("[UploadSweeper] 收集文件引用失败，本轮跳过: %v", err)
		return 0
	}

	cutoff := time.Now().Add(-orphanGrace)
	removed := 0
	removedDirs := map[string]bool{}
	_ = filepath.WalkDir("upload", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil // 目录与不可读项跳过，不影响其余遍历
		}
		info, statErr := d.Info()
		if statErr != nil || info.ModTime().After(cutoff) {
			return nil // 宽限期内的新文件不动（可能正在等课程保存）
		}
		url := "/" + filepath.ToSlash(path)
		if refs[url] {
			return nil
		}
		if os.Remove(path) == nil {
			removed++
			removedDirs[filepath.Dir(path)] = true
		}
		return nil
	})

	// 顺手清理因删除而变空的叶子目录（os.Remove 对非空目录报错，天然安全）
	for dir := range removedDirs {
		for dir != "upload" && strings.HasPrefix(dir, "upload") {
			if os.Remove(dir) != nil {
				break
			}
			dir = filepath.Dir(dir)
		}
	}
	return removed
}

// collectUploadRefs 全量收集仍在引用上传文件的 URL 集合。
// 软删记录（课程/视频/学员等）被 Model 默认作用域过滤，其文件视为可清理。
func collectUploadRefs() (map[string]bool, error) {
	refs := map[string]bool{}
	refCols := []struct {
		model  interface{}
		column string
	}{
		{&adminModel.Video{}, "url"},
		{&adminModel.Video{}, "thumbnail"},
		{&adminModel.Video{}, "courseware"},
		{&adminModel.Video{}, "courseware_pdf"},
		{&adminModel.Course{}, "cover"},
		{&adminModel.Class{}, "cover"},
		{&adminModel.Admin{}, "avatar"},
		{&userModel.User{}, "avatar"},
	}
	for _, c := range refCols {
		var arr []string
		if err := global.DB.Model(c.model).Pluck("DISTINCT "+c.column, &arr).Error; err != nil {
			return nil, err
		}
		for _, u := range arr {
			if u != "" {
				refs[u] = true
			}
		}
	}
	// settings.value 可能内嵌 URL（如 JSON 配置里的 logo），正则提取保护
	var vals []string
	if err := global.DB.Model(&adminModel.Setting{}).Pluck("value", &vals).Error; err != nil {
		return nil, err
	}
	for _, v := range vals {
		for _, u := range uploadURLRe.FindAllString(v, -1) {
			refs[u] = true
		}
	}
	return refs, nil
}
