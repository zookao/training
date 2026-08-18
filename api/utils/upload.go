package utils

import (
	"crypto/md5"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// allowedVideoExts 允许的视频格式（仅限浏览器 <video> 原生可播放的容器格式）
//
// AVI/MKV/MOV/FLV/WMV/M4V/MPG/MPEG/TS/3GP 等容器浏览器无法原生播放，
// 即便 ffmpeg 可处理，学员端也无法播放，故一律拒绝。
var allowedVideoExts = map[string]bool{
	".mp4":  true, // MP4（H.264/AAC，所有主流浏览器支持）
	".webm": true, // WebM（VP8/VP9，Chrome/Firefox/Edge/Safari 14+ 支持）
}

// allowedCoursewareExts 课件允许的文件扩展名（LibreOffice 可转 PDF 的演示文稿格式）
var allowedCoursewareExts = map[string]bool{
	".pptx": true, // PowerPoint 2007+
	".ppt":  true, // PowerPoint 97-2003
	".odp":  true, // OpenDocument Presentation
	".fodp": true, // Flat OpenDocument Presentation
}

// IsAllowedVideoExt 判断文件名是否为允许的视频格式
func IsAllowedVideoExt(filename string) bool {
	return allowedVideoExts[strings.ToLower(filepath.Ext(filename))]
}

// IsAllowedCoursewareExt 判断文件名是否为允许的课件格式
func IsAllowedCoursewareExt(filename string) bool {
	return allowedCoursewareExts[strings.ToLower(filepath.Ext(filename))]
}

// VideoFinalPath 根据原始文件名计算视频最终落地的相对路径与 url（不创建文件）。
// 用于分片合并写入目标，与 FinalizeVideo 共享同一套命名规则。
func VideoFinalPath(filename string) (relPath, url string, err error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" || !allowedVideoExts[ext] {
		return "", "", errors.New("不支持的视频格式，仅支持 MP4/WebM（浏览器可播放格式）")
	}
	date := time.Now().Format("060102") // 6位日期 YYMMDD
	nameHash := md5sum(filename)
	videoFileName := nameHash + ext
	relPath = filepath.Join("upload", date, "videos", videoFileName)
	url = fmt.Sprintf("/upload/%s/videos/%s", date, videoFileName)
	return relPath, url, nil
}

// CoursewareFinalPath 根据原始文件名计算课件最终落地的相对路径与 url（不创建文件）。
func CoursewareFinalPath(filename string) (relPath, url string, err error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedCoursewareExts[ext] {
		return "", "", errors.New("仅支持 PPTX/PPT/ODP/FODP 格式课件")
	}
	date := time.Now().Format("060102")
	nameHash := md5sum(filename)
	fileName := nameHash + ext
	relPath = filepath.Join("upload", date, "pptx", fileName)
	url = fmt.Sprintf("/upload/%s/pptx/%s", date, fileName)
	return relPath, url, nil
}

// dateFromPath 从 upload/<date>/... 相对路径解析出 date（YYMMDD）。
// 兼容路径分隔符差异（Windows \ / Unix /）。
func dateFromPath(relPath string) (string, error) {
	norm := strings.ReplaceAll(relPath, "\\", "/")
	parts := strings.Split(norm, "/")
	// 期望形如 ["upload", "250727", "videos", "<file>"]
	if len(parts) < 3 || parts[0] != "upload" {
		return "", fmt.Errorf("无法从路径解析日期: %s", relPath)
	}
	return parts[1], nil
}

// FinalizeVideo 对已落地的视频文件执行后处理（ffprobe 时长 + 缩略图）。
//   - mergedLocalPath: 已合并/已保存的视频相对路径（如 upload/250727/videos/<hash>.mp4）
//   - filename:        原始文件名（派生 url 中的 hash + 扩展名校验）
//   - thumbnail:       可选用户上传缩略图（nil 则 ffmpeg 自动截取第 10% 位置的帧）
//
// date 从 mergedLocalPath 路径解析（不取 time.Now()，避免跨午夜时 url 与文件实际路径漂移）。
// 本函数不删除 mergedLocalPath，失败清理由调用方负责。
func FinalizeVideo(c *gin.Context, mergedLocalPath, filename string, thumbnail *multipart.FileHeader) (url, thumbURL string, duration int, err error) {
	videoExt := strings.ToLower(filepath.Ext(filename))
	if videoExt == "" || !allowedVideoExts[videoExt] {
		return "", "", 0, errors.New("不支持的视频格式，仅支持 MP4/WebM（浏览器可播放格式）")
	}
	date, err := dateFromPath(mergedLocalPath)
	if err != nil {
		return "", "", 0, err
	}
	nameHash := md5sum(filename)
	videoFileName := nameHash + videoExt
	url = fmt.Sprintf("/upload/%s/videos/%s", date, videoFileName)

	// 1. ffprobe 时长
	videoAbs, _ := filepath.Abs(mergedLocalPath)
	duration, derr := GetVideoDuration(videoAbs)
	if derr != nil {
		return "", "", 0, fmt.Errorf("获取视频时长失败: %w", derr)
	}

	// 2. 缩略图：用户上传则用上传的；未上传则用 ffmpeg 自动截取
	thumbDir := filepath.Join("upload", date, "thumbnails")
	if err = os.MkdirAll(thumbDir, 0755); err != nil {
		return url, "", duration, err
	}

	if thumbnail != nil {
		thumbExt := strings.ToLower(filepath.Ext(thumbnail.Filename))
		if thumbExt == "" {
			thumbExt = ".jpg"
		}
		thumbFileName := nameHash + thumbExt
		thumbDst := filepath.Join(thumbDir, thumbFileName)
		if err = c.SaveUploadedFile(thumbnail, thumbDst); err != nil {
			return url, "", duration, err
		}
		thumbURL = fmt.Sprintf("/upload/%s/thumbnails/%s", date, thumbFileName)
	} else {
		thumbFileName := nameHash + ".jpg"
		thumbDst := filepath.Join(thumbDir, thumbFileName)
		if eerr := ExtractThumbnail(videoAbs, thumbDst); eerr != nil {
			// 截取失败不阻断流程，返回空缩略图
			return url, "", duration, nil
		}
		thumbURL = fmt.Sprintf("/upload/%s/thumbnails/%s", date, thumbFileName)
	}
	return url, thumbURL, duration, nil
}

// SaveVideo 保存视频文件（及可选缩略图）到 upload/<YYMMDD>/videos|thumbnails/<md5(文件名)>.<ext>
// 返回视频 url、缩略图 url、视频时长（秒）。
//
// 旧的单次上传入口（保留向后兼容）；分片上传走 FinalizeVideo。
func SaveVideo(c *gin.Context, video, thumbnail *multipart.FileHeader) (url, thumbURL string, duration int, err error) {
	videoDst, _, err := VideoFinalPath(video.Filename)
	if err != nil {
		return "", "", 0, err
	}
	if err = os.MkdirAll(filepath.Dir(videoDst), 0755); err != nil {
		return "", "", 0, err
	}
	if err = c.SaveUploadedFile(video, videoDst); err != nil {
		return "", "", 0, err
	}
	url, thumbURL, duration, err = FinalizeVideo(c, videoDst, video.Filename, thumbnail)
	if err != nil {
		_ = os.Remove(videoDst) // 失败清理（保持旧行为）
		return "", "", 0, err
	}
	return url, thumbURL, duration, nil
}

// SaveImage 保存图片文件（课程/班级封面等）到 upload/<YYMMDD>/images/<md5(文件名)>.<ext>
// 返回图片 url。文件名取 md5(原始文件名) + 原始扩展名。
func SaveImage(c *gin.Context, image *multipart.FileHeader) (url string, err error) {
	date := time.Now().Format("060102") // 6位日期 YYMMDD
	nameHash := md5sum(image.Filename)
	// 图片扩展名：取原始扩展名，缺省补 .jpg
	imgExt := strings.ToLower(filepath.Ext(image.Filename))
	if imgExt == "" {
		imgExt = ".jpg"
	}
	fileName := nameHash + imgExt

	dir := filepath.Join("upload", date, "images")
	if err = os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, fileName)
	if err = c.SaveUploadedFile(image, dst); err != nil {
		return "", err
	}
	url = fmt.Sprintf("/upload/%s/images/%s", date, fileName)
	return url, nil
}

// FinalizeCourseware 对已落地的课件文件执行后处理（LibreOffice 转 PDF + 页数统计）。
//   - mergedLocalPath: 已合并/已保存的课件相对路径（如 upload/250727/pptx/<hash>.pptx）
//   - filename:        原始文件名（派生 url 中的 hash + 扩展名校验）
//
// date 从 mergedLocalPath 路径解析。本函数不删除 mergedLocalPath，失败清理由调用方负责。
func FinalizeCourseware(mergedLocalPath, filename string) (coursewareURL, pdfURL string, pageCount int, err error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedCoursewareExts[ext] {
		return "", "", 0, errors.New("仅支持 PPTX/PPT/ODP/FODP 格式课件")
	}
	date, err := dateFromPath(mergedLocalPath)
	if err != nil {
		return "", "", 0, err
	}
	nameHash := md5sum(filename)
	fileName := nameHash + ext
	coursewareURL = fmt.Sprintf("/upload/%s/pptx/%s", date, fileName)

	// 调用 LibreOffice 转 PDF（需要绝对路径）
	cwAbs, _ := filepath.Abs(mergedLocalPath)
	pdfDir := filepath.Join("upload", date, "pdf")
	pdfAbsDir, _ := filepath.Abs(pdfDir)
	pdfAbsPath, convErr := ConvertToPDF(cwAbs, pdfAbsDir)
	if convErr != nil {
		return "", "", 0, fmt.Errorf("课件转 PDF 失败: %w", convErr)
	}
	pdfURL = fmt.Sprintf("/upload/%s/pdf/%s", date, filepath.Base(pdfAbsPath))

	// 从 PDF 统计页数（比从原始文件统计更通用，支持所有格式）
	pageCount, _ = CountPDFPages(pdfAbsPath)
	return coursewareURL, pdfURL, pageCount, nil
}

// SaveCourseware 保存课件文件并转换为 PDF 用于在线预览。
// 旧的单次上传入口（保留向后兼容）；分片上传走 FinalizeCourseware。
func SaveCourseware(c *gin.Context, file *multipart.FileHeader) (coursewareURL, pdfURL string, pageCount int, err error) {
	cwDst, _, err := CoursewareFinalPath(file.Filename)
	if err != nil {
		return "", "", 0, err
	}
	if err = os.MkdirAll(filepath.Dir(cwDst), 0755); err != nil {
		return "", "", 0, err
	}
	if err = c.SaveUploadedFile(file, cwDst); err != nil {
		return "", "", 0, err
	}
	coursewareURL, pdfURL, pageCount, err = FinalizeCourseware(cwDst, file.Filename)
	if err != nil {
		_ = os.Remove(cwDst) // 转换失败：清理已保存的课件文件，硬性要求不降级
		return "", "", 0, err
	}
	return coursewareURL, pdfURL, pageCount, nil
}

// DeleteUploadByURL 根据 /upload/... URL 删除本地文件（不存在则忽略）
func DeleteUploadByURL(url string) {
	local, ok := URLToLocalPath(url)
	if !ok {
		return
	}
	_ = os.Remove(local)
}

// URLToLocalPath 将 /upload/... URL 转换为本地相对路径
// 返回 (路径, 是否有效)
func URLToLocalPath(url string) (string, bool) {
	if url == "" || !strings.HasPrefix(url, "/upload/") {
		return "", false
	}
	rel := strings.TrimPrefix(url, "/upload/")
	return filepath.FromSlash(filepath.Join("upload", rel)), true
}

func md5sum(s string) string {
	h := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", h)
}
