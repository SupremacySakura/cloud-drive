package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "cloud-drive-backend/docs"
)

func SetUpRouter() *gin.Engine {
	r := gin.Default()
	RegisterUserRouter(r)
	RegisterFileRouter(r)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}
