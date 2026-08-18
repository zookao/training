package middleware

import (
	"strings"

	"training/global"
	"training/model/common"
	"training/utils"

	"github.com/gin-gonic/gin"
)

const (
	ctxClaimsKey = "auth_claims"
	ctxTokenKey  = "auth_token"
)

// JWTAuth JWT 鉴权中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			// 兼容静态资源（video/img 标签无法设置 Header）：从 query param 读取
			token = c.Query("token")
		}
		if token == "" {
			common.FailWithCode(c, common.CodeNoAuth, "未登录或登录已过期")
			c.Abort()
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		// 检查 token 是否已被注销（黑名单）
		if utils.IsTokenBlacklisted(token) {
			common.FailWithCode(c, common.CodeNoAuth, "登录已失效，请重新登录")
			c.Abort()
			return
		}
		claims, err := utils.ParseToken(token, global.Config.JWT.SigningKey)
		if err != nil {
			common.FailWithCode(c, common.CodeNoAuth, "登录凭证无效")
			c.Abort()
			return
		}
		c.Set(ctxClaimsKey, claims)
		c.Set(ctxTokenKey, token)
		c.Next()
	}
}

// GuardType 限制 guard 类型（admin/user）
func GuardType(guard string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			common.FailWithCode(c, common.CodeNoAuth, "未登录")
			c.Abort()
			return
		}
		if claims.GuardType != guard {
			common.FailWithCode(c, common.CodeNoPerm, "无权限访问该端")
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetClaims 从上下文获取 claims
func GetClaims(c *gin.Context) *utils.CustomClaims {
	v, ok := c.Get(ctxClaimsKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*utils.CustomClaims)
	return claims
}

// GetToken 从上下文获取原始 token 字符串
func GetToken(c *gin.Context) string {
	v, ok := c.Get(ctxTokenKey)
	if !ok {
		return ""
	}
	token, _ := v.(string)
	return token
}
