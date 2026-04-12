package handler

import (
	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/middleware"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/response"
	"cloud-drive-backend/internal/service"
	"cloud-drive-backend/internal/vo"
	"math/rand"
	"strings"

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
	r.GET("/list", middleware.AuthMiddleware(), h.GetListByFolderIDAndUserID)
	r.GET("/list/count", middleware.AuthMiddleware(), h.GetListCountByFolderIDAndUserID)
	r.POST("/mkdir", middleware.AuthMiddleware(), h.MakeDirectory)
	r.POST("/code", middleware.AuthMiddleware(), h.CreatePickUpCode)
	r.GET("/code/list", middleware.AuthMiddleware(), h.GetPickUpCodeListByUserIDAndPage)
	r.GET("/code/count", middleware.AuthMiddleware(), h.GetPickUpCodeListCountByUserID)
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

// GetListByFolderIDAndUserID godoc
// @Summary 获取文件夹下的文件列表
// @Description 获取文件夹下的文件列表
// @Tags file
// @Accept json
// @Produce json
// @Param folder_id query int true "文件夹ID"
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} response.Response{data=[]dto.FileListItem} "成功返回（code=0,data=[]dto.FileListItem）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/list [get]
func (h *FileHandler) GetListByFolderIDAndUserID(c *gin.Context) {
	var req dto.GetListByFolderIDAndUserIDReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	list, err := h.FileService.GetListByFolderIDAndUserID(req.FolderID, userID, req.Page, req.PageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, list)
}

// GetListCountByFolderIDAndUserID godoc
// @Summary 获取文件夹下的文件数量
// @Description 获取文件夹下的文件数量
// @Tags file
// @Accept json
// @Produce json
// @Param folder_id query int true "文件夹ID"
// @Success 200 {object} response.Response{data=int64} "成功返回（code=0,data=int64）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/list/count [get]
func (h *FileHandler) GetListCountByFolderIDAndUserID(c *gin.Context) {
	var req dto.GetListCountByFolderIDAndUserIDReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	count, err := h.FileService.GetListCountByFolderIDAndUserID(req.FolderID, userID)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, count)
}

func (h *FileHandler) MakeDirectory(c *gin.Context) {
	var req dto.MakeDirectoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	id, err := h.FileService.MakeDirectory(req.FolderID, req.Name, userID)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, id)
}

// CreatePickUpCode godoc
// @Summary 创建分享码
// @Description 创建分享码
// @Tags file
// @Accept json
// @Produce json
// @Param data body dto.CreatePickUpCodeReq true "分享码参数"
// @Success 200 {object} response.Response{data=int} "成功返回（code=0,data=int）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/code [post]
func (h *FileHandler) CreatePickUpCode(c *gin.Context) {
	var req dto.CreatePickUpCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	code := req.Code
	if code == "" {
		code = generatePickUpCode()
	}
	code, ok = normalizePickUpCode(code)
	if !ok {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	id, err := h.FileService.CreatePickUpCode(&model.PickUpCodeModel{
		Code:        code,
		Status:      model.PickUpCodeStatusActive,
		UserID:      userID,
		Type:        req.Type,
		FileID:      req.FileID,
		FolderID:    req.FolderID,
		MaxDownload: uint(req.MaxDownloads),
		ExpireTime:  req.ExpireTime,
	})
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, id)
}

func generatePickUpCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func normalizePickUpCode(code string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	normalized = strings.ReplaceAll(normalized, "-", "")
	if len(normalized) != 6 {
		return "", false
	}
	for _, ch := range normalized {
		if !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') {
			return "", false
		}
	}
	return normalized, true
}

func (h *FileHandler) GetPickUpCodeListByUserIDAndPage(c *gin.Context) {
	var req dto.GetPickUpCodeListByUserIDAndPageReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	list, err := h.FileService.GetPickUpCodeListByUserID(userID, req.Page, req.PageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, list)
}

func (h *FileHandler) GetPickUpCodeListCountByUserID(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	count, err := h.FileService.GetPickUpCodeListCountByUserID(userID)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, count)
}
