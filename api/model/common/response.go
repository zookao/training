package common

import (
	"training/utils"

	"github.com/gin-gonic/gin"
)

// Result 统一响应结构
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess = 0
	CodeFail    = 1
	CodeNoAuth  = 401
	CodeNoPerm  = 403
)

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	utils.RenderJSON(c, 200, Result{Code: CodeSuccess, Msg: "success", Data: data})
}

// OKMsg 成功响应带消息
func OKMsg(c *gin.Context, msg string) {
	utils.RenderJSON(c, 200, Result{Code: CodeSuccess, Msg: msg})
}

// Fail 失败响应
func Fail(c *gin.Context, msg string) {
	utils.RenderJSON(c, 200, Result{Code: CodeFail, Msg: msg})
}

// FailWithCode 失败响应带状态码
func FailWithCode(c *gin.Context, code int, msg string) {
	utils.RenderJSON(c, 200, Result{Code: code, Msg: msg})
}
