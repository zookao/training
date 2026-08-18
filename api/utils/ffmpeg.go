package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"training/global"
)

// ffmpegTimeout 单次 ffmpeg/ffprobe 调用超时
// 探测时长 + 抽取缩略帧，大视频可能较慢，留 5 分钟
const ffmpegTimeout = 300 * time.Second

// FindFFmpeg 查找 ffmpeg 可执行文件路径。
// 优先级：config.yaml 的 ffmpeg.path > PATH 中的 ffmpeg
// 找不到返回错误（硬性要求 ffmpeg，不降级）。
func FindFFmpeg() (string, error) {
	// 1. 配置指定的路径
	if p := global.Config.FFmpeg.Path; p != "" {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
		return "", fmt.Errorf("配置的 ffmpeg.path 不可用: %s", p)
	}
	// 2. PATH 探测 ffmpeg
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path, nil
	}
	return "", errors.New("未找到 ffmpeg，请在 config.yaml 配置 ffmpeg.path，或安装 ffmpeg（macOS: brew install ffmpeg / Linux: apt install ffmpeg）")
}

// FindFFprobe 查找 ffprobe 可执行文件路径（与 ffmpeg 同目录）。
func FindFFprobe() (string, error) {
	ffmpegPath, err := FindFFmpeg()
	if err != nil {
		return "", err
	}
	// 1. 同目录下的 ffprobe
	dir := filepath.Dir(ffmpegPath)
	ffprobeCandidate := filepath.Join(dir, "ffprobe")
	if path, err := exec.LookPath(ffprobeCandidate); err == nil {
		return path, nil
	}
	// 2. PATH 中的 ffprobe
	if path, err := exec.LookPath("ffprobe"); err == nil {
		return path, nil
	}
	return "", errors.New("未找到 ffprobe（通常与 ffmpeg 一起安装）")
}

// GetVideoDuration 使用 ffprobe 获取视频时长（秒，向下取整）。
func GetVideoDuration(videoPath string) (int, error) {
	ffprobe, err := FindFFprobe()
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), ffmpegTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, ffprobe,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe 获取时长失败: %w", err)
	}
	durStr := strings.TrimSpace(string(out))
	if durStr == "" || durStr == "N/A" {
		return 0, fmt.Errorf("无法获取视频时长，文件可能已损坏")
	}
	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		return 0, fmt.Errorf("解析时长失败: %w", err)
	}
	return int(dur), nil
}

// ExtractThumbnail 使用 ffmpeg 从视频中截取一帧作为封面图。
//   - 截取位置：视频时长的 10%（至少 1 秒），避免黑屏片头
//   - 输出格式：JPG，质量 q:v=2
//   - outPath: 输出文件路径（需调用方确保目录存在）
func ExtractThumbnail(videoPath, outPath string) error {
	ffmpeg, err := FindFFmpeg()
	if err != nil {
		return err
	}
	// 决定截取位置
	seekSec := 1
	if dur, derr := GetVideoDuration(videoPath); derr == nil && dur > 0 {
		seekSec = dur / 10
		if seekSec < 1 {
			seekSec = 1
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), ffmpegTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, ffmpeg,
		"-ss", fmt.Sprintf("%d", seekSec),
		"-i", videoPath,
		"-frames:v", "1",
		"-q:v", "2",
		"-y",
		outPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg 截取封面失败: %w, 输出: %s", err, string(out))
	}
	if _, err := os.Stat(outPath); err != nil {
		return fmt.Errorf("ffmpeg 截取封面后未找到输出文件: %w", err)
	}
	return nil
}
