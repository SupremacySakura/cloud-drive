package handler

import (
	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/response"
	"cloud-drive-backend/internal/service"
	"cloud-drive-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

func (h *AuthHandler) Register(r *gin.RouterGroup) {
	r.POST("/register", h.RegisterUser)
	r.POST("/login", h.Login)
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	// 注册用户
	var req dto.RegisterUserReq
	// 绑定 JSON 请求体到 req 结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	// 密码哈希
	hashPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	// 创建新用户
	newUser := &model.UserModel{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashPassword,
		AvatarURL:    "",
		StorageLimit: 1024 * 1024 * 1024, // 1GB
		StorageUsed:  0,
		Status:       1,
	}
	// 注册用户
	if err := h.AuthService.RegisterUser(newUser); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	// 返回成功响应
	response.Success(c, nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	// 登录用户
	var req dto.LoginReq
	// 绑定 JSON 请求体到 req 结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	// 验证用户
	user, err := h.AuthService.ValidateUser(req.Username, req.Password)
	if err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}
	// 生成 JWT 令牌
	token, err := h.AuthService.GenerateToken(user.ID)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	// 返回成功响应
	response.Success(c, gin.H{
		"token": token,
	})
}
