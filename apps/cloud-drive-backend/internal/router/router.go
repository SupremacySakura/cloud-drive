package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "cloud-drive-backend/docs"
	"cloud-drive-backend/internal/service"
)

func SetUpRouter() *gin.Engine {
	r := gin.Default()
	RegisterUserRouter(r)
	RegisterFileRouter(r, service.FileServiceOptions{
		ChunkStoragePath: "/Users/shi/study/frontend/projects/cloud-drive/data",
		FileStoragePath:  "/Users/shi/study/frontend/projects/cloud-drive/data",
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}
