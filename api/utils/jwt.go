package utils

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims JWT 自定义声明
type CustomClaims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	GuardType string `json:"guardType"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT
func GenerateToken(userID uint, username, guardType, signingKey, issuer string, expiresHours int) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		GuardType: guardType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        strconv.FormatInt(time.Now().UnixNano(), 36), // jti: 每次生成唯一，避免同秒内登出再登录产生相同 token 被黑名单误杀
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey))
}

// ParseToken 解析 JWT
func ParseToken(tokenStr, signingKey string) (*CustomClaims, error) {
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
