package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"cloud-drive-backend/internal/log"
)

// LoggerMiddleware 请求日志中间件，记录 method、path、status、latency、client_ip
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 获取状态码
		statusCode := c.Writer.Status()

		// 如果有查询参数，附加到路径
		if raw != "" {
			path = path + "?" + raw
		}

		// 记录结构化日志
		log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("client_ip", clientIP).
			Msg("HTTP请求")
	}
}
