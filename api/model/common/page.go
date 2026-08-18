package common

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PageRequest 分页请求
type PageRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	Keyword  string `form:"keyword" json:"keyword"`
}

// PageList 分页响应
type PageList struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// Normalize 分页参数默认值
func (p *PageRequest) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 200 {
		p.PageSize = 20
	}
}

// Offset 计算偏移
func (p *PageRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// ParseIDParam 从 URL 路径参数解析 uint ID，失败时返回 0 和错误
func ParseIDParam(c *gin.Context, key string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || id == 0 {
		return 0, errInvalidID
	}
	return uint(id), nil
}

// ParseQueryID 从 query 参数解析 uint ID（未传或无效时返回 0）
func ParseQueryID(c *gin.Context, key string) uint {
	id, err := strconv.ParseUint(c.Query(key), 10, 64)
	if err != nil || id == 0 {
		return 0
	}
	return uint(id)
}

var errInvalidID = &invalidIDError{}

type invalidIDError struct{}

func (e *invalidIDError) Error() string { return "无效的ID参数" }
