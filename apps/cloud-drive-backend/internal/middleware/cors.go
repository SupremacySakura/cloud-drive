package middleware

import (
    "time"

    "cloud-drive-backend/internal/config"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

// CORSMiddleware 根据配置配置 CORS 策略
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
    origins := cfg.GetCORSOrigins()

    config := cors.Config{
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }

    if len(origins) == 0 {
        config.AllowAllOrigins = true
    } else {
        config.AllowOrigins = origins
    }

    return cors.New(config)
}
