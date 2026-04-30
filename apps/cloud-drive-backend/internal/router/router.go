package router

import (
    "net/http"

    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"

    _ "cloud-drive-backend/docs"
    "cloud-drive-backend/internal/config"
    "cloud-drive-backend/internal/service"
    "cloud-drive-backend/internal/middleware"
)

func SetUpRouter(cfg *config.Config) *gin.Engine {
    r := gin.Default()
    // 注册跨域、限流和日志中间件
    r.Use(middleware.CORSMiddleware(cfg))
    r.Use(middleware.RateLimitMiddleware())
    r.Use(middleware.LoggerMiddleware())
	RegisterUserRouter(r)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.HEAD("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	RegisterFileRouter(r, service.FileServiceOptions{
		ChunkStoragePath: cfg.ChunkStoragePath,
		FileStoragePath:  cfg.FileStoragePath,
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}
