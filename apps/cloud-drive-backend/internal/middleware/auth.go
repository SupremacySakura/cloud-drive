package middleware

import (
	"cloud-drive-backend/internal/response"
	"cloud-drive-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(int(response.CodeUnauthorized), gin.H{"error": "未授权"})
			c.Abort()
			return
		}
		// 验证token
		claims, err := utils.ParseToken(token[7:])
		if err != nil {
			c.JSON(int(response.CodeUnauthorized), gin.H{"error": "无效的token"})
			c.Abort()
			return
		}
		// 如果token有效，设置用户ID到上下文
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
