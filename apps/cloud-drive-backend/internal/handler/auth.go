package handler

import (
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthService *service.AuthService
}

func (h AuthHandler) Register(r *gin.RouterGroup) {
	r.POST("/register", h.RegisterUser)
}

func (h AuthHandler) RegisterUser(c *gin.Context) {
	var req model.UserModel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.AuthService.RegisterUser(&req); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "register success",
	})
}
