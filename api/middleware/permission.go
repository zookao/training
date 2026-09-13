package middleware

import (
	"sync"
	"time"

	"training/global"
	"training/model/admin"
	"training/model/common"

	"github.com/gin-gonic/gin"
)

// permCacheTTL 管理员接口权限缓存时长：角色/权限变更最迟该时长后生效。
// 低频运维操作换 QPS，可接受；如需即时生效可调小
const permCacheTTL = 60 * time.Second

type permCacheEntry struct {
	perms    map[string]bool
	expireAt time.Time
}

var (
	permCacheMu sync.RWMutex
	permCache   = map[uint]permCacheEntry{}
)

// adminApiPerms 查询并缓存管理员可访问的 "METHOD /path" 集合。
// 原实现每个管理端请求都三表 join 查库，是管理端吞吐的第一热点；
// 这里按管理员缓存 60s（条目按 adminID 覆盖写，无泄漏风险）
func adminApiPerms(adminID uint) (map[string]bool, error) {
	permCacheMu.RLock()
	e, ok := permCache[adminID]
	permCacheMu.RUnlock()
	if ok && time.Now().Before(e.expireAt) {
		return e.perms, nil
	}

	var rows []struct {
		Method string
		Path   string
	}
	if err := global.DB.Model(&admin.Api{}).
		Joins("JOIN role_apis ON role_apis.api_id = apis.id").
		Joins("JOIN admin_roles ON admin_roles.role_id = role_apis.role_id").
		Where("admin_roles.admin_id = ?", adminID).
		Select("DISTINCT apis.method, apis.path").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	perms := make(map[string]bool, len(rows))
	for _, r := range rows {
		perms[r.Method+" "+r.Path] = true
	}
	permCacheMu.Lock()
	permCache[adminID] = permCacheEntry{perms: perms, expireAt: time.Now().Add(permCacheTTL)}
	permCacheMu.Unlock()
	return perms, nil
}

// HasApiPerm 校验当前管理员是否拥有访问该接口的权限
// 通过 admin→roles→apis 关系匹配 method+path
func HasApiPerm() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			common.FailWithCode(c, common.CodeNoAuth, "未登录")
			c.Abort()
			return
		}
		// 超管（id=1）直接放行
		if claims.UserID == 1 {
			c.Next()
			return
		}
		method := c.Request.Method
		path := c.FullPath() // 使用路由模板，如 /api/admin/admin/:id
		if path == "" {
			path = c.Request.URL.Path
		}

		perms, err := adminApiPerms(claims.UserID)
		if err != nil {
			// 缓存未命中且查询失败：退回逐请求查库的老路径，不因缓存故障拒绝服务
			var count int64
			global.DB.Model(&admin.Api{}).
				Joins("JOIN role_apis ON role_apis.api_id = apis.id").
				Joins("JOIN admin_roles ON admin_roles.role_id = role_apis.role_id").
				Where("admin_roles.admin_id = ? AND apis.method = ? AND apis.path = ?", claims.UserID, method, path).
				Count(&count)
			if count == 0 {
				common.FailWithCode(c, common.CodeNoPerm, "无接口访问权限")
				c.Abort()
				return
			}
			c.Next()
			return
		}
		if !perms[method+" "+path] {
			common.FailWithCode(c, common.CodeNoPerm, "无接口访问权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
