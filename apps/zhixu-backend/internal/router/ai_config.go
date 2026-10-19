package router

import (
	"zhixu-backend/internal/handler"
	"zhixu-backend/internal/middleware"
	"zhixu-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterAIConfigRouter(r *gin.Engine, svc *service.AIConfigService) {
	h := &handler.AIConfigHandler{Service: svc}
	group := r.Group("/settings/ai", middleware.AuthMiddleware())
	group.GET("", h.Get)
	group.GET("/catalog", h.Catalog)
	group.POST("/providers", h.Create)
	group.PUT("/providers/:id", h.Update)
	group.DELETE("/providers/:id", h.Delete)
	group.PUT("/defaults", h.Defaults)
}
