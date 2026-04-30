package router

import (
    "cloud-drive-backend/internal/database"
    "cloud-drive-backend/internal/handler"
    "cloud-drive-backend/internal/repository"
    "cloud-drive-backend/internal/service"

    "github.com/gin-gonic/gin"
)

// 全局 FileService 单例（DI 使用）
var fileServiceSingleton service.FileService

func RegisterFileRouter(r *gin.Engine, options service.FileServiceOptions) {
    fileGroup := r.Group("/file")
    // 初始化数据库服务
    FileRepository := repository.NewFileRepository(database.DB)
    UserRepository := repository.NewUserRepository(database.DB)
    // 依赖注入单例：文件服务在进程内单例化
    if fileServiceSingleton == nil {
        fileServiceSingleton = service.NewFileService(FileRepository, options)
    }
    AuthService := service.NewAuthService(UserRepository)
    // 初始化控制器
    FileHandler := handler.NewFileHandler(fileServiceSingleton, AuthService)
    // 注册控制器
    FileHandler.Register(fileGroup)
}
