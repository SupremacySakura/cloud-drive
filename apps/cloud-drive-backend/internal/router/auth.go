package router

import (
	"cloud-drive-backend/internal/database"
	"cloud-drive-backend/internal/handler"
	"cloud-drive-backend/internal/repository"
	"cloud-drive-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterUserRouter(r *gin.Engine) {
	userGroup := r.Group("/auth")
	// 初始化数据库服务
	UserRepository := repository.NewUserRepository(database.DB)
	// 初始化业务服务
	AuthService := service.NewAuthService(UserRepository)
	// 初始化控制器
	AuthHandler := handler.NewAuthHandler(AuthService)
	// 注册控制器
	AuthHandler.Register(userGroup)
}
