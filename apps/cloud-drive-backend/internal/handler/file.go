package handler

import (
	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/middleware"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/response"
	"cloud-drive-backend/internal/service"
	"cloud-drive-backend/internal/vo"

	"github.com/gin-gonic/gin"
)

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

type FileHandler struct {
	FileService service.FileService
	AuthService service.AuthService
}

func NewFileHandler(fileService service.FileService, authService service.AuthService) *FileHandler {
	return &FileHandler{
		FileService: fileService,
		AuthService: authService,
	}
}

func (h *FileHandler) Register(r *gin.RouterGroup) {
	r.POST("/init", middleware.AuthMiddleware(), h.InitUploadFile)
	r.POST("/chunk", middleware.AuthMiddleware(), h.UploadFileChunk)
}

// InitUploadFile godoc
// @Summary 初始化文件分片上传
// @Description 初始化文件分片上传任务，返回已上传分片列表（同hash任务会复用）
// @Tags file
// @Accept json
// @Produce json
// @Param data body dto.InitUploadFileReq true "初始化上传参数"
// @Success 200 {object} response.Response{data=vo.InitUploadFileResp} "成功返回（code=0,data=vo.InitUploadFileResp）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/init [post]
func (h *FileHandler) InitUploadFile(c *gin.Context) {
	var req dto.InitUploadFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	// 从上下文获取用户ID
	user_id, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	// 转换为uint类型
	userID, ok := user_id.(uint)
	if !ok {
		response.Fail(c, response.CodeServerError)
		return
	}
	_, err := h.AuthService.GetUserByID(userID)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	task := &model.UploadTask{
		FileName:       req.FileName,
		FileSize:       uint64(req.FileSize),
		FileHash:       req.FileHash,
		ChunkSize:      req.ChunkSize,
		TotalChunks:    req.TotalChunks,
		UploadedChunks: model.IntSlice{},
		FileType:       req.FileType,
		FolderID:       req.FolderID,
		UserID:         userID,
	}
	task, err = h.FileService.InitUploadFile(task)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, vo.InitUploadFileResp{
		UploadedChunks: []int(task.UploadedChunks),
	})
}

// UploadFileChunk godoc
// @Summary 上传文件分片
// @Description 上传文件分片（当前接口实现为占位逻辑，后续补齐分片参数解析与落库/合并）
// @Tags file
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "成功返回（code=0）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/chunk [post]
func (h *FileHandler) UploadFileChunk(c *gin.Context) {
	if err := h.FileService.UploadFileChunk(); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, nil)
}
