package router

import (
	"cloud-drive-backend/internal/database"
	"cloud-drive-backend/internal/handler"
	"cloud-drive-backend/internal/repository"
	"cloud-drive-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterFileRouter(r *gin.Engine, options service.FileServiceOptions) {
	fileGroup := r.Group("/file")
	// 初始化数据库服务
	FileRepository := repository.NewFileRepository(database.DB)
	UserRepository := repository.NewUserRepository(database.DB)
	// 初始化业务服务
	FileService := service.NewFileService(FileRepository, options)
	AuthService := service.NewAuthService(UserRepository)
	// 初始化控制器
	FileHandler := handler.NewFileHandler(FileService, AuthService)
	// 注册控制器
	FileHandler.Register(fileGroup)
}
