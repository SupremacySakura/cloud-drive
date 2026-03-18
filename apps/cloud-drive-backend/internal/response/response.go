package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ===== 错误码类型 =====
type ResponseCode int

// ===== 错误码定义 =====
const (
	CodeSuccess      ResponseCode = 0
	CodeInvalidParam ResponseCode = 1001
	CodeUnauthorized ResponseCode = 1002
	CodeNotFound     ResponseCode = 1003
	CodeServerError  ResponseCode = 1004
)

// ===== code -> msg 映射（简单版）=====
func (c ResponseCode) Msg() string {
	switch c {
	case CodeSuccess:
		return "success"
	case CodeInvalidParam:
		return "参数错误"
	case CodeUnauthorized:
		return "未登录"
	case CodeNotFound:
		return "资源不存在"
	case CodeServerError:
		return "服务器错误"
	default:
		return "未知错误"
	}
}

// ===== 统一返回结构 =====
type Response struct {
	Code ResponseCode `json:"code"`
	Msg  string       `json:"msg"`
	Data interface{}  `json:"data,omitempty"`
}

// ===== 成功 =====
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}

// ===== 失败 =====
func Fail(c *gin.Context, code ResponseCode) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  code.Msg(),
	})
}

// ===== 自定义错误信息 =====
func FailWithMsg(c *gin.Context, code ResponseCode, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}
