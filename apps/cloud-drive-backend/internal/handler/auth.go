package handler

import (
	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/middleware"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/response"
	"cloud-drive-backend/internal/service"
	"cloud-drive-backend/internal/utils"
	"cloud-drive-backend/internal/vo"

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
	r.GET("/check", middleware.AuthMiddleware(), h.checkLogin)
}

// RegisterUser godoc
// @Summary 用户注册
// @Description 用户通过用户名、邮箱、密码注册
// @Tags auth
// @Accept json
// @Produce json
// @Param data body dto.RegisterUserReq true "注册参数"
// @Success 200 {object} response.Response{data=nil} "返回统一响应（code=0成功，code!=0失败；data可能为空）"
// @Router /auth/register [post]
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

// Login godoc
// @Summary 用户登录
// @Description 用户通过账号密码登录
// @Tags auth
// @Accept json
// @Produce json
// @Param data body dto.LoginReq true "登录参数"
// @Success 200 {object} response.Response{data=vo.LoginResp} "返回统一响应（code=0成功，code!=0失败；data包含token）"
// @Router /auth/login [post]
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
	response.Success(c, vo.LoginResp{
		Token: token,
	})
}

func (h *AuthHandler) checkLogin(c *gin.Context) {
	// 从上下文获取用户ID
	_, exists := c.Get("user_id")
	if !exists {
		response.FailWithMsg(c, response.CodeUnauthorized, "unauthorized")
		return
	}
	response.Success(c, nil)
}
