package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
设计决策说明：业务码模式与HTTP状态码的共存

本项目采用混合模式：
1. 保留业务码(code字段) - 用于前端精确识别错误类型，便于国际化和精细化处理
2. 新增HTTP状态码 - 用于HTTP层面的语义化，便于API网关、监控、缓存等基础设施识别

为什么保留业务码而不是完全改为RESTful状态码？
1. 向后兼容性：前端已大量依赖code字段，完全移除改动成本过高
2. 精细化控制：业务码能表达更细致的错误类型（如1001参数错误可细分字段）
3. 跨层一致性：业务码在日志、监控、错误追踪中保持一致性
4. 渐进式演进：可以逐步改进而无需一次性重构所有前端代码

使用建议：
- 新增API优先使用FailWithStatus返回语义化HTTP状态码
- 保留业务码用于详细的错误分类
- handler层负责映射业务错误到HTTP状态码
- service层专注于业务错误，不关心HTTP细节
*/

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

// ===== 带HTTP状态码的错误返回 =====
// 注意：虽然返回指定的HTTP状态码，但仍保留业务码模式
// 这是为了渐进式改进API，保持向后兼容性
func FailWithStatus(c *gin.Context, httpStatus int, code ResponseCode, msg string) {
	c.JSON(httpStatus, Response{
		Code: code,
		Msg:  msg,
	})
}
