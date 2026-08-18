package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"training/global"
)

// convertTimeout 单次 PPTX→PDF 转换的超时时间
// 大型 PPTX（含图表/动画/嵌入对象）转换较慢，留足 10 分钟避免误杀
const convertTimeout = 600 * time.Second

// queueTimeout 排队等待超时：前方任务过多时，最多等待 N 秒后返回错误
// 默认 30 分钟（3 个 10 分钟的大文件转换排队是合理上限），可通过 config.yaml libreoffice.queueTimeoutSec 配置
var queueTimeout = 30 * time.Minute

// ---- LibreOffice 并发转换信号量（任务队列） ----
// LibreOffice headless 并发不稳定（共享字体缓存/资源竞争，Linux Docker 下尤为明显），
// 用 buffered channel 做信号量限制并发数：容量 N = 最多同时 N 个转换，
// 超出的请求自动排队等待（FIFO），实现"自动排队任务等待"。
var (
	convertSemOnce sync.Once
	convertSem     chan struct{}
)

// getConvertSem 懒初始化信号量（需等 config 加载完成后才能读取配置）
func getConvertSem() chan struct{} {
	convertSemOnce.Do(func() {
		n := 1 // 默认严格串行，最安全
		if global.Config != nil && global.Config.LibreOffice.MaxConcurrency > 0 {
			n = global.Config.LibreOffice.MaxConcurrency
		}
		convertSem = make(chan struct{}, n)
		// 排队超时也可配置（默认 30 分钟）
		if global.Config != nil && global.Config.LibreOffice.QueueTimeoutSec > 0 {
			queueTimeout = time.Duration(global.Config.LibreOffice.QueueTimeoutSec) * time.Second
		}
		log.Printf("[LibreOffice] 并发转换限制: %d, 排队超时: %v", n, queueTimeout)
	})
	return convertSem
}

// FindSoffice 查找 soffice 可执行文件路径。
// 优先级：config.yaml 的 libreoffice.path > PATH 中的 soffice > PATH 中的 libreoffice
// 找不到返回错误（硬性要求 LibreOffice，不降级）。
func FindSoffice() (string, error) {
	// 1. 配置指定的路径
	if p := global.Config.LibreOffice.Path; p != "" {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
		// 配置了但找不到，直接报错（避免静默回退掩盖配置错误）
		return "", fmt.Errorf("配置的 libreoffice.path 不可用: %s", p)
	}
	// 2. PATH 探测 soffice
	if path, err := exec.LookPath("soffice"); err == nil {
		return path, nil
	}
	// 3. PATH 探测 libreoffice
	if path, err := exec.LookPath("libreoffice"); err == nil {
		return path, nil
	}
	return "", errors.New("未找到 LibreOffice，请在 config.yaml 配置 libreoffice.path，或安装 LibreOffice（macOS: brew install --cask libreoffice / Linux: apt install libreoffice）")
}

// ConvertToPDF 调用 LibreOffice 把演示文稿（PPTX/PPT/ODP/FODP 等）转成 PDF。
//   - inputPath: 输入文件路径（绝对路径）
//   - outDir:    输出目录（绝对路径），不存在会创建
//
// 返回生成的 PDF 绝对路径。文件名规则：输入 <name>.<ext> → 输出 <name>.pdf
//
// 并发安全：
//  1. 信号量排队：多用户同时上传课件时，按 config libreoffice.maxConcurrency 限制并发数（默认 1），
//     超出的请求自动排队等待，避免 LibreOffice 并发转换失败。
//  2. 独立 profile：每次调用用独立的 UserInstallation 目录，避免 profile 冲突。
func ConvertToPDF(inputPath, outDir string) (string, error) {
	soffice, err := FindSoffice()
	if err != nil {
		return "", err
	}

	// 输出目录确保存在
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 独立的 user profile 目录，避免并发冲突
	profileDir, err := os.MkdirTemp("", "lo-profile-")
	if err != nil {
		return "", fmt.Errorf("创建临时 profile 目录失败: %w", err)
	}
	defer os.RemoveAll(profileDir)
	// file:// URI（profileDir 是绝对路径）
	profileURI := "file://" + profileDir

	// 输出 PDF 路径：与输入同名，扩展名换 .pdf
	base := filepath.Base(inputPath)
	pdfName := strings.TrimSuffix(base, filepath.Ext(base)) + ".pdf"
	pdfPath := filepath.Join(outDir, pdfName)

	// 已存在同名 PDF 先删除，避免 LibreOffice 跳过转换
	_ = os.Remove(pdfPath)

	// 排队等待信号量：LibreOffice headless 并发不稳定，限制并发数
	// 超出限制的请求在此阻塞等待，实现自动排队
	sem := getConvertSem()
	select {
	case sem <- struct{}{}:
		// 成功获取信号量，转换完成后释放
		defer func() { <-sem }()
		log.Printf("[LibreOffice] 开始转换: %s", base)
	case <-time.After(queueTimeout):
		return "", errors.New("课件转换排队超时（前方任务较多），请稍后重试")
	}

	ctx, cancel := context.WithTimeout(context.Background(), convertTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, soffice,
		"--headless",
		"--norestore",
		"--nologo",
		"--nolockcheck",
		"--convert-to", "pdf",
		"--outdir", outDir,
		"-env:UserInstallation="+profileURI,
		inputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("LibreOffice 转换失败: %w, 输出: %s", err, string(out))
	}
	// 校验输出文件
	if _, err := os.Stat(pdfPath); err != nil {
		return "", fmt.Errorf("LibreOffice 转换后未找到 PDF 文件: %w, 输出: %s", err, string(out))
	}
	log.Printf("[LibreOffice] 转换完成: %s → %s", base, pdfName)
	return pdfPath, nil
}
