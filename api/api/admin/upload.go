package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"training/middleware"
	adminModel "training/model/admin"
	"training/model/common"
	"training/utils"

	"github.com/gin-gonic/gin"
)

// 分片上传：init（查已传）→ chunk（传分片）→ merge（流式合并 + 后处理 + 清理）
//
// 分片暂存于 .chunks/<adminId>/<uploadId>/<chunkIndex>（位于二进制工作目录，
// 不在 upload/ 下，避免被 /upload/ 静态 token 服务暴露）。
// init 时写 .meta JSON，chunk/merge 据此交叉校验参数，不盲信客户端。

const (
	chunkStaleTTL = 24 * time.Hour // 分片会话过期清理阈值
	chunkMaxSize  = 32 << 20       // 单分片硬上限 32MB（正常 5MB）
)

var uploadIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// chunkMeta 分片上传元信息（落盘到 .chunks/<adminId>/<uploadId>/.meta）
type chunkMeta struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	TotalChunks int    `json:"totalChunks"`
	ChunkSize   int64  `json:"chunkSize"`
	Type        string `json:"type"` // "video" | "courseware"
	CreatedAt   int64  `json:"createdAt"`
}

func chunksRoot() string { return ".chunks" }

func adminChunkDir(adminID uint, uploadID string) string {
	return filepath.Join(chunksRoot(), strconv.FormatUint(uint64(adminID), 10), uploadID)
}

func readChunkMeta(dir string) (*chunkMeta, error) {
	data, err := os.ReadFile(filepath.Join(dir, ".meta"))
	if err != nil {
		return nil, err
	}
	var m chunkMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func writeChunkMeta(dir string, m *chunkMeta) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".meta"), data, 0644)
}

// listUploadedChunks 列出分片目录中已存在的分片序号（升序）
func listUploadedChunks(dir string) []int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []int{}
	}
	var indices []int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == ".meta" {
			continue
		}
		idx, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	return indices
}

// staleSweep 扫描某管理员下所有分片会话，删除过期（.meta.createdAt 超过 chunkStaleTTL）的目录
func staleSweep(adminDir string) {
	entries, err := os.ReadDir(adminDir)
	if err != nil {
		return
	}
	now := time.Now()
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sub := filepath.Join(adminDir, e.Name())
		m, err := readChunkMeta(sub)
		if err != nil {
			continue
		}
		if time.Unix(m.CreatedAt, 0).Add(chunkStaleTTL).Before(now) {
			_ = os.RemoveAll(sub)
		}
	}
}

// UploadChunkInit POST /api/admin/upload/chunk/init
// 初始化分片上传会话，返回已上传的分片序号（用于断点续传）
func UploadChunkInit(c *gin.Context) {
	var req struct {
		UploadID    string `json:"uploadId"`
		Filename    string `json:"filename"`
		Size        int64  `json:"size"`
		TotalChunks int    `json:"totalChunks"`
		ChunkSize   int64  `json:"chunkSize"`
		Type        string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误: "+err.Error())
		return
	}
	if !uploadIDRe.MatchString(req.UploadID) {
		common.Fail(c, "uploadId 非法")
		return
	}
	if req.Type != "video" && req.Type != "courseware" {
		common.Fail(c, "type 必须为 video 或 courseware")
		return
	}
	if req.TotalChunks < 1 || req.ChunkSize < 1 || req.Size < 1 {
		common.Fail(c, "参数非法")
		return
	}
	// totalChunks 与 size/chunkSize 一致性校验
	expected := (req.Size + req.ChunkSize - 1) / req.ChunkSize
	if req.TotalChunks != int(expected) {
		common.Fail(c, "totalChunks 与 size/chunkSize 不一致")
		return
	}
	// 扩展名校验
	if req.Type == "video" && !utils.IsAllowedVideoExt(req.Filename) {
		common.Fail(c, "不支持的视频格式，仅支持 MP4/WebM（浏览器可播放格式）")
		return
	}
	if req.Type == "courseware" && !utils.IsAllowedCoursewareExt(req.Filename) {
		common.Fail(c, "仅支持 PPTX/PPT/ODP/FODP 格式课件")
		return
	}

	claims := middleware.GetClaims(c)
	if claims == nil {
		common.FailWithCode(c, common.CodeNoAuth, "未登录")
		return
	}
	adminDir := filepath.Join(chunksRoot(), strconv.FormatUint(uint64(claims.UserID), 10))
	staleSweep(adminDir) // 清理过期会话

	dir := adminChunkDir(claims.UserID, req.UploadID)
	// 已存在会话？
	if m, err := readChunkMeta(dir); err == nil {
		// 交叉校验参数一致（防止客户端篡改）
		if m.Filename != req.Filename || m.Size != req.Size ||
			m.TotalChunks != req.TotalChunks || m.ChunkSize != req.ChunkSize || m.Type != req.Type {
			common.Fail(c, "上传参数与已存在会话不一致，请更换文件或清空后重试")
			return
		}
		common.OK(c, gin.H{"uploaded": listUploadedChunks(dir), "totalChunks": m.TotalChunks})
		return
	}

	// 全新会话
	if err := os.MkdirAll(dir, 0755); err != nil {
		common.Fail(c, "创建分片目录失败: "+err.Error())
		return
	}
	m := &chunkMeta{
		Filename:    req.Filename,
		Size:        req.Size,
		TotalChunks: req.TotalChunks,
		ChunkSize:   req.ChunkSize,
		Type:        req.Type,
		CreatedAt:   time.Now().Unix(),
	}
	if err := writeChunkMeta(dir, m); err != nil {
		common.Fail(c, "写入元信息失败: "+err.Error())
		return
	}
	common.OK(c, gin.H{"uploaded": []int{}, "totalChunks": req.TotalChunks})
}

// UploadChunk POST /api/admin/upload/chunk
// 上传单个分片（幂等：同序号覆盖）
func UploadChunk(c *gin.Context) {
	uploadID := c.PostForm("uploadId")
	if !uploadIDRe.MatchString(uploadID) {
		common.Fail(c, "uploadId 非法")
		return
	}
	chunkIndex, err := strconv.Atoi(c.PostForm("chunkIndex"))
	if err != nil {
		common.Fail(c, "chunkIndex 非法")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		common.Fail(c, "缺少分片文件")
		return
	}

	claims := middleware.GetClaims(c)
	if claims == nil {
		common.FailWithCode(c, common.CodeNoAuth, "未登录")
		return
	}
	dir := adminChunkDir(claims.UserID, uploadID)
	m, err := readChunkMeta(dir)
	if err != nil {
		common.Fail(c, "上传会话不存在，请先初始化")
		return
	}
	if chunkIndex < 0 || chunkIndex >= m.TotalChunks {
		common.Fail(c, "chunkIndex 越界")
		return
	}
	// 分片大小：绝不超过声明的 chunkSize（末片可小），且不超硬上限
	if file.Size > m.ChunkSize {
		common.Fail(c, "分片大小超过声明值")
		return
	}
	if file.Size > chunkMaxSize {
		common.Fail(c, "分片过大")
		return
	}

	dst := filepath.Join(dir, strconv.Itoa(chunkIndex))
	if err := c.SaveUploadedFile(file, dst); err != nil {
		common.Fail(c, "保存分片失败: "+err.Error())
		return
	}
	common.OK(c, gin.H{"ok": true, "index": chunkIndex})
}

// UploadChunkMerge POST /api/admin/upload/chunk/merge
// 流式合并所有分片到最终文件，执行后处理（ffprobe/缩略图 或 LibreOffice/PDF），清理分片目录
func UploadChunkMerge(c *gin.Context) {
	uploadID := c.PostForm("uploadId")
	if !uploadIDRe.MatchString(uploadID) {
		common.Fail(c, "uploadId 非法")
		return
	}
	filename := c.PostForm("filename")
	fileType := c.PostForm("type")
	totalChunks, err := strconv.Atoi(c.PostForm("totalChunks"))
	if err != nil {
		common.Fail(c, "totalChunks 非法")
		return
	}
	size, err := strconv.ParseInt(c.PostForm("size"), 10, 64)
	if err != nil {
		common.Fail(c, "size 非法")
		return
	}
	if fileType != "video" && fileType != "courseware" {
		common.Fail(c, "type 必须为 video 或 courseware")
		return
	}

	claims := middleware.GetClaims(c)
	if claims == nil {
		common.FailWithCode(c, common.CodeNoAuth, "未登录")
		return
	}
	dir := adminChunkDir(claims.UserID, uploadID)
	m, err := readChunkMeta(dir)
	if err != nil {
		common.Fail(c, "上传会话不存在，请先初始化")
		return
	}
	// 交叉校验参数与会话一致
	if m.Filename != filename || m.Size != size || m.TotalChunks != totalChunks || m.Type != fileType {
		common.Fail(c, "合并参数与会话不一致")
		return
	}
	// 扩展名校验
	if fileType == "video" && !utils.IsAllowedVideoExt(filename) {
		common.Fail(c, "不支持的视频格式")
		return
	}
	if fileType == "courseware" && !utils.IsAllowedCoursewareExt(filename) {
		common.Fail(c, "仅支持 PPTX/PPT/ODP/FODP 格式课件")
		return
	}

	// 最终落地路径
	var relPath string
	if fileType == "video" {
		relPath, _, err = utils.VideoFinalPath(filename)
	} else {
		relPath, _, err = utils.CoursewareFinalPath(filename)
	}
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	if err := os.MkdirAll(filepath.Dir(relPath), 0755); err != nil {
		common.Fail(c, "创建目标目录失败: "+err.Error())
		return
	}

	// 流式合并（O(1) 内存）
	mergedFile, err := os.OpenFile(relPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		common.Fail(c, "创建合并文件失败: "+err.Error())
		return
	}
	for i := 0; i < m.TotalChunks; i++ {
		chunkPath := filepath.Join(dir, strconv.Itoa(i))
		cf, err := os.Open(chunkPath)
		if err != nil {
			mergedFile.Close()
			_ = os.Remove(relPath)
			common.Fail(c, fmt.Sprintf("分片 %d 缺失，请续传后重试", i))
			return
		}
		if _, err := io.Copy(mergedFile, cf); err != nil {
			cf.Close()
			mergedFile.Close()
			_ = os.Remove(relPath)
			common.Fail(c, fmt.Sprintf("合并分片 %d 失败: %s", i, err.Error()))
			return
		}
		cf.Close()
	}
	mergedFile.Close()

	// 校验合并后大小（防分片损坏/截断）
	info, err := os.Stat(relPath)
	if err != nil || info.Size() != m.Size {
		_ = os.Remove(relPath)
		common.Fail(c, "合并后大小校验失败，请续传后重试")
		return
	}

	// 后处理
	if fileType == "video" {
		thumbnail, _ := c.FormFile("thumbnail") // 可选
		url, thumbURL, duration, ferr := utils.FinalizeVideo(c, relPath, filename, thumbnail)
		if ferr != nil {
			_ = os.Remove(relPath)
			common.Fail(c, "视频后处理失败: "+ferr.Error())
			return
		}
		_ = os.RemoveAll(dir) // 成功：清理分片
		common.OK(c, adminModel.UploadVideoRes{
			URL:       url,
			Thumbnail: thumbURL,
			Filename:  filename,
			Duration:  duration,
		})
		return
	}

	// courseware
	cwURL, pdfURL, pageCount, ferr := utils.FinalizeCourseware(relPath, filename)
	if ferr != nil {
		_ = os.Remove(relPath)
		common.Fail(c, ferr.Error())
		return
	}
	_ = os.RemoveAll(dir)
	common.OK(c, adminModel.UploadVideoRes{
		Courseware:          cwURL,
		CoursewarePDF:       pdfURL,
		CoursewarePageCount: pageCount,
	})
}
