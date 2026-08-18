package utils

import (
	"net/http"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// CST 东八区（UTC+8），用于统一时间输出，不依赖系统时区与 tzdata
var CST = time.FixedZone("CST", 8*3600)

// TimeLayout 系统统一时间格式：2023-08-01 12:00:00
const TimeLayout = "2006-01-02 15:04:05"

// jsonAPI 全局 JSON 序列化器：与标准库行为兼容；time.Time 的输出格式由下方
// 包级 RegisterTypeEncoderFunc 覆盖为东八区 "2006-01-02 15:04:05"。
var jsonAPI jsoniter.API = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	// time.Time -> 东八区 "2006-01-02 15:04:05"（零值输出空字符串）
	// *time.Time 由 jsoniter 自动解引用后复用此编码器，nil 输出 null。
	jsoniter.RegisterTypeEncoderFunc("time.Time", encodeTime, nil)
}

func encodeTime(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	t := *(*time.Time)(ptr)
	if t.IsZero() {
		stream.WriteString("")
		return
	}
	stream.WriteString(t.In(CST).Format(TimeLayout))
}

// RenderJSON 使用全局序列化器输出 JSON，统一应用时间格式。
func RenderJSON(c *gin.Context, status int, v interface{}) {
	c.Render(status, jsonRender{Data: v})
}

// jsonRender 基于 jsoniter 的 gin 渲染器
type jsonRender struct {
	Data interface{}
}

func (r jsonRender) Render(w http.ResponseWriter) error {
	writeJSONContentType(w)
	buf, err := jsonAPI.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

func (r jsonRender) WriteContentType(w http.ResponseWriter) {
	writeJSONContentType(w)
}

func writeJSONContentType(w http.ResponseWriter) {
	header := w.Header()
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", "application/json; charset=utf-8")
	}
}
