package utils

import (
	"sync"
	"time"
)

// tokenBlacklist 内存级 JWT 黑名单，用于 Logout 注销。
// key: 原始 token 字符串；value: token 过期时间（到期后自动清理）。
type tokenBlacklist struct {
	mu   sync.RWMutex
	data map[string]time.Time
	stop chan struct{}
}

var blacklist = &tokenBlacklist{
	data: make(map[string]time.Time),
	stop: make(chan struct{}),
}

// startCleanup 启动后台清理协程，定期移除已过期条目
func init() {
	go blacklist.cleanup()
}

// cleanup 每 10 分钟清理一次已过期的黑名单条目
func (b *tokenBlacklist) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			now := time.Now()
			for token, exp := range b.data {
				if now.After(exp) {
					delete(b.data, token)
				}
			}
			b.mu.Unlock()
		case <-b.stop:
			return
		}
	}
}

// BlacklistToken 将 token 加入黑名单，直到 expiry 过期后自动移除
func BlacklistToken(token string, expiry time.Time) {
	if token == "" {
		return
	}
	blacklist.mu.Lock()
	blacklist.data[token] = expiry
	blacklist.mu.Unlock()
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func IsTokenBlacklisted(token string) bool {
	if token == "" {
		return false
	}
	blacklist.mu.RLock()
	_, ok := blacklist.data[token]
	blacklist.mu.RUnlock()
	return ok
}
