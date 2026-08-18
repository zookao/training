package config

import (
	"os"

	"github.com/spf13/viper"
)

// Server HTTP 服务配置
type Server struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// MySQL 数据库配置
type MySQL struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"maxIdleConns"`
	MaxOpenConns int    `mapstructure:"maxOpenConns"`
}

// JWT 配置
type JWT struct {
	SigningKey  string `mapstructure:"signingKey"`
	ExpiresTime int    `mapstructure:"expiresTime"`
	Issuer      string `mapstructure:"issuer"`
}

// LibreOffice 配置（PPTX 转 PDF 用）
type LibreOffice struct {
	Path               string `mapstructure:"path"`               // soffice 二进制路径，空则探测 PATH
	MaxConcurrency     int    `mapstructure:"maxConcurrency"`     // 最大并发转换数，默认 1（串行排队），LibreOffice headless 并发不稳定
	QueueTimeoutSec    int    `mapstructure:"queueTimeoutSec"`    // 排队等待超时（秒），默认 1800=30分钟，超时返回错误提示用户稍后重试
}

// FFmpeg 配置（视频处理用：获取时长、截取封面）
type FFmpeg struct {
	Path string `mapstructure:"path"` // ffmpeg 二进制路径，空则探测 PATH（ffprobe 需同目录）
}

// Config 全局配置
type Config struct {
	Server      Server      `mapstructure:"server"`
	MySQL       MySQL       `mapstructure:"mysql"`
	JWT         JWT         `mapstructure:"jwt"`
	LibreOffice LibreOffice `mapstructure:"libreoffice"`
	FFmpeg      FFmpeg      `mapstructure:"ffmpeg"`
}

// Load 加载配置文件
// 敏感字段（JWT 签名密钥、MySQL 密码）支持环境变量覆盖，优先级：环境变量 > config.yaml
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	// 环境变量覆盖（生产环境推荐通过环境变量注入敏感配置）
	if envKey := os.Getenv("JWT_SIGNING_KEY"); envKey != "" {
		v.Set("jwt.signingKey", envKey)
	}
	if envDBPwd := os.Getenv("MYSQL_PASSWORD"); envDBPwd != "" {
		v.Set("mysql.password", envDBPwd)
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
