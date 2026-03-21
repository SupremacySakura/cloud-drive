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
	r.POST("/merge", middleware.AuthMiddleware(), h.MergeUploadedChunks)
}

func getCurrentUserID(c *gin.Context) (uint, bool) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		return 0, false
	}
	return userID, true
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
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
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
		TaskID:         task.ID,
		UploadedChunks: []int(task.UploadedChunks),
		Status:         task.Status,
	})
}

// UploadFileChunk godoc
// @Summary 上传文件分片
// @Description 上传文件分片
// @Tags file
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "成功返回（code=0）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/chunk [post]
func (h *FileHandler) UploadFileChunk(c *gin.Context) {
	var req dto.UploadChunkReq

	// 1. 绑定普通字段（form）
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	// 2. 获取用户
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	// 3. 获取文件 chunk_data（form-data）
	file, err := c.FormFile("chunk_data")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	// 4. 打开文件流
	src, err := file.Open()
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	defer src.Close()

	// 5. 交给 service（传 io.Reader，而不是 base64）
	if err := h.FileService.UploadFileChunkStream(userID, &req, src); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}

	response.Success(c, nil)
}

// MergeUploadedChunks godoc
// @Summary 合并上传的文件分片
// @Description 合并上传的文件分片
// @Tags file
// @Accept json
// @Produce json
// @Param data body dto.MergeUploadedChunksReq true "合并上传参数"
// @Success 200 {object} response.Response "成功返回（code=0）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/merge [post]
func (h *FileHandler) MergeUploadedChunks(c *gin.Context) {
	var req dto.MergeUploadedChunksReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	if err := h.FileService.MergeUploadedChunks(userID, req.TaskID); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, nil)
}