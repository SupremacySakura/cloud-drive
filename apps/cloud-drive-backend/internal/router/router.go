package router

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "cloud-drive-backend/docs"
	"cloud-drive-backend/internal/service"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SetUpRouter() *gin.Engine {
	r := gin.Default()
	RegisterUserRouter(r)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.HEAD("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 从环境变量读取存储路径，默认为项目目录下的 data 文件夹
	chunkPath := getEnvOrDefault("CHUNK_STORAGE_PATH", "./data")
	filePath := getEnvOrDefault("FILE_STORAGE_PATH", "./data")

	RegisterFileRouter(r, service.FileServiceOptions{
		ChunkStoragePath: chunkPath,
		FileStoragePath:  filePath,
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}
