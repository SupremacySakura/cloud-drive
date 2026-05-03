package handler

import (
	"bytes"
	"crypto/rand"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"

	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/errors"
	"cloud-drive-backend/internal/middleware"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/response"
	"cloud-drive-backend/internal/service"
	"cloud-drive-backend/internal/vo"

	"github.com/gin-gonic/gin"
)

// 文件大小限制常量
const (
	maxUploadFileSize uint64 = 100 * 1024 * 1024
	maxChunkSize      int64  = 10 * 1024 * 1024
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
	r.GET("/dashboard/overview", middleware.AuthMiddleware(), h.GetDashboardOverview)
	r.GET("/list", middleware.AuthMiddleware(), h.GetListByFolderIDAndUserID)
	r.GET("/list/count", middleware.AuthMiddleware(), h.GetListCountByFolderIDAndUserID)
	r.GET("/preview", middleware.AuthMiddleware(), h.PreviewFileByID)
	r.GET("/download", middleware.AuthMiddleware(), h.DownloadFileByID)
	r.POST("/share/link", middleware.AuthMiddleware(), h.CreatePublicShareLink)
	r.GET("/share/link", middleware.AuthMiddleware(), h.GetPublicShareLink)
	r.DELETE("/share/link", middleware.AuthMiddleware(), h.DeletePublicShareLink)
	r.GET("/share/open", h.OpenPublicShare)
	r.POST("/mkdir", middleware.AuthMiddleware(), h.MakeDirectory)
	r.POST("/rename", middleware.AuthMiddleware(), h.RenameFileByID)
	r.POST("/move", middleware.AuthMiddleware(), h.MoveFileByID)
	r.POST("/delete", middleware.AuthMiddleware(), h.DeleteFileByID)
	r.POST("/code", middleware.AuthMiddleware(), h.CreatePickUpCode)
	r.GET("/code/list", middleware.AuthMiddleware(), h.GetPickUpCodeListByUserIDAndPage)
	r.GET("/code/count", middleware.AuthMiddleware(), h.GetPickUpCodeListCountByUserID)
	r.DELETE("/code", middleware.AuthMiddleware(), h.DeletePickUpCodeByID)
	r.GET("/pickup/download", h.DownloadByPickUpCode)
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

	if strings.Contains(req.FileHash, "..") || strings.Contains(req.FileHash, "/") || strings.Contains(req.FileHash, "\\") {
		response.FailWithMsg(c, response.CodeInvalidParam, "无效的文件哈希")
		return
	}

	if req.FileSize > maxUploadFileSize {
		response.FailWithMsg(c, response.CodeInvalidParam, "文件大小超过限制")
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

	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	file, err := c.FormFile("chunk_data")
	if err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	if file.Size > maxChunkSize {
		response.FailWithMsg(c, response.CodeInvalidParam, "分片大小超过限制")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	defer src.Close()

	limitedReader := io.LimitReader(src, maxChunkSize+1)
	buf := make([]byte, 512)
	n, err := limitedReader.Read(buf)
	if err != nil && err != io.EOF {
		response.Fail(c, response.CodeServerError)
		return
	}

	detectedMIME := http.DetectContentType(buf[:n])
	if !h.FileService.IsAllowedMIMEType(detectedMIME) {
		response.FailWithMsg(c, response.CodeInvalidParam, "不支持的文件类型")
		return
	}

	remainingData, err := io.ReadAll(limitedReader)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}

	if int64(len(remainingData)) > maxChunkSize-int64(n) {
		response.FailWithMsg(c, response.CodeInvalidParam, "分片大小超过限制")
		return
	}

	fullData := append(buf[:n], remainingData...)
	reader := bytes.NewReader(fullData)

	if err := h.FileService.UploadFileChunkStream(userID, &req, reader, int64(len(fullData))); err != nil {
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

// GetDashboardOverview godoc
// @Summary 获取仪表盘概览数据
// @Description 获取当前用户的空间使用、文件类型统计、最近活动
// @Tags file
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.DashboardOverviewResp} "成功返回（code=0,data=dto.DashboardOverviewResp）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/dashboard/overview [get]
func (h *FileHandler) GetDashboardOverview(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	user, err := h.AuthService.GetUserByID(userID)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	overview, err := h.FileService.GetDashboardOverview(userID, user.StorageLimit)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.Success(c, overview)
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

// PreviewFileByID godoc
// @Summary 预览文件
// @Description 通过文件ID预览文件内容（内联显示）
// @Tags file
// @Accept json
// @Produce octet-stream
// @Param file_id query int true "文件ID"
// @Success 200 {file} binary "文件内容流"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Failure 403 {object} map[string]string "权限不足"
// @Failure 404 {object} map[string]string "文件不存在"
// @Security ApiKeyAuth
// @Router /file/preview [get]
func (h *FileHandler) PreviewFileByID(c *gin.Context) {
	var req dto.PreviewFileReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	setMeta := func(fileName, contentType string) {
		escapedName := url.PathEscape(fileName)
		c.Header("Content-Description", "File Preview")
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", "inline; filename*=UTF-8''"+escapedName)
		c.Header("Cache-Control", "no-cache")
	}
	if err := h.FileService.PreviewFileByID(req.FileID, userID, c.Writer, setMeta); err != nil {
		if errors.Is(err, errors.ErrFileNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在")
			return
		}
		if errors.Is(err, errors.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在或无权访问")
		return
	}
}

// DownloadFileByID godoc
// @Summary 下载文件或文件夹
// @Description 通过 file_id 下载单文件，或通过 folder_id 下载文件夹（zip）
// @Tags file
// @Accept json
// @Produce octet-stream
// @Param file_id query int false "文件ID"
// @Param folder_id query int false "文件夹ID"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/download [get]
func (h *FileHandler) DownloadFileByID(c *gin.Context) {
	var req dto.DownloadFileReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	if (req.FileID == 0 && req.FolderID == 0) || (req.FileID > 0 && req.FolderID > 0) {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	setMeta := func(fileName, contentType string) {
		escapedName := url.PathEscape(fileName)
		c.Header("Content-Description", "File Download")
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+escapedName)
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Cache-Control", "no-cache")
	}
	if err := h.FileService.DownloadByIDs(userID, req.FileID, req.FolderID, c.Writer, setMeta); err != nil {
		if errors.Is(err, errors.ErrPickupEmptyFolder) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件夹为空")
			return
		}
		if errors.Is(err, errors.ErrFileNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在")
			return
		}
		if errors.Is(err, errors.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在或无权访问")
		return
	}
}

// CreatePublicShareLink godoc
// @Summary 创建公开分享链接
// @Description 为文件创建公开分享链接，生成分享token
// @Tags file
// @Accept json
// @Produce json
// @Param data body dto.CreatePublicShareLinkReq true "创建分享参数"
// @Success 200 {object} response.Response{data=map[string]string} "成功返回（code=0,data包含token和url）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Failure 403 {object} map[string]string "权限不足"
// @Failure 404 {object} map[string]string "文件不存在"
// @Security ApiKeyAuth
// @Router /file/share/link [post]
func (h *FileHandler) CreatePublicShareLink(c *gin.Context) {
	var req dto.CreatePublicShareLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	token, err := h.FileService.CreatePublicShareLink(req.FileID, userID)
	if err != nil {
		if errors.Is(err, errors.ErrFileNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在")
			return
		}
		if errors.Is(err, errors.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在或无权访问")
		return
	}
	shareURL := buildPublicShareURL(c, token)
	response.Success(c, gin.H{
		"token": token,
		"url":   shareURL,
	})
}

// GetPublicShareLink godoc
// @Summary 获取公开分享链接
// @Description 获取文件的公开分享链接信息
// @Tags file
// @Accept json
// @Produce json
// @Param file_id query int true "文件ID"
// @Success 200 {object} response.Response{data=map[string]interface{}} "成功返回（code=0,data包含exists、token、url）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Failure 403 {object} map[string]string "权限不足"
// @Failure 404 {object} map[string]string "文件不存在"
// @Security ApiKeyAuth
// @Router /file/share/link [get]
func (h *FileHandler) GetPublicShareLink(c *gin.Context) {
	var req dto.GetPublicShareLinkReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	token, err := h.FileService.GetPublicShareLink(req.FileID, userID)
	if err != nil {
		if errors.Is(err, errors.ErrPublicShareNotFound) {
			response.Success(c, gin.H{
				"exists": false,
				"token":  "",
				"url":    "",
			})
			return
		}
		if errors.Is(err, errors.ErrFileNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在")
			return
		}
		if errors.Is(err, errors.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在或无权访问")
		return
	}
	response.Success(c, gin.H{
		"exists": true,
		"token":  token,
		"url":    buildPublicShareURL(c, token),
	})
}

// DeletePublicShareLink godoc
// @Summary 删除公开分享链接
// @Description 删除文件的公开分享链接
// @Tags file
// @Accept json
// @Produce json
// @Param file_id query int true "文件ID"
// @Success 200 {object} response.Response "成功返回（code=0）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Failure 403 {object} map[string]string "权限不足"
// @Failure 404 {object} map[string]string "分享链接不存在或文件不存在"
// @Security ApiKeyAuth
// @Router /file/share/link [delete]
func (h *FileHandler) DeletePublicShareLink(c *gin.Context) {
	var req dto.DeletePublicShareLinkReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	if err := h.FileService.DeletePublicShareLink(req.FileID, userID); err != nil {
		if errors.Is(err, errors.ErrPublicShareNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "分享链接不存在")
			return
		}
		if errors.Is(err, errors.ErrFileNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在")
			return
		}
		if errors.Is(err, errors.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在或无权访问")
		return
	}
	response.Success(c, nil)
}

func buildPublicShareURL(c *gin.Context, token string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			scheme = strings.TrimSpace(parts[0])
		}
	}
	return scheme + "://" + c.Request.Host + "/file/share/open?token=" + url.QueryEscape(token)
}

// OpenPublicShare godoc
// @Summary 打开公开分享
// @Description 通过分享token访问公开分享的文件（无需登录）
// @Tags file
// @Accept json
// @Produce octet-stream
// @Param token query string true "分享token"
// @Success 200 {file} binary "文件内容流"
// @Failure 404 {object} map[string]string "分享链接不存在或文件已删除"
// @Router /file/share/open [get]
func (h *FileHandler) OpenPublicShare(c *gin.Context) {
	var req dto.OpenPublicShareReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	setMeta := func(fileName, contentType string) {
		escapedName := url.PathEscape(fileName)
		c.Header("Content-Description", "Public Share")
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", "inline; filename*=UTF-8''"+escapedName)
		c.Header("Cache-Control", "public, max-age=300")
	}
	if err := h.FileService.OpenPublicShare(req.Token, c.Writer, setMeta); err != nil {
		if errors.Is(err, errors.ErrPublicShareNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "分享链接不存在或文件已删除")
			return
		}
		response.FailWithStatus(c, http.StatusInternalServerError, response.CodeServerError, "服务器错误")
		return
	}
}

// MakeDirectory godoc
// @Summary 创建文件夹
// @Description 在指定文件夹下创建新文件夹
// @Tags file
// @Accept json
// @Produce json
// @Param data body dto.MakeDirectoryReq true "创建文件夹参数"
// @Success 200 {object} response.Response{data=uint} "成功返回（code=0,data=文件夹ID）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/mkdir [post]
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

// RenameFileByID godoc
// @Summary 重命名文件或文件夹
// @Description 通过 file_id 或 folder_id 重命名（两者必须且只能传一个）
// @Tags file
// @Accept json
// @Produce json
// @Param data body dto.RenameFileReq true "重命名参数"
// @Success 200 {object} response.Response "成功返回（code=0）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/rename [post]
func (h *FileHandler) RenameFileByID(c *gin.Context) {
	var req dto.RenameFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	if (req.FileID == 0 && req.FolderID == 0) || (req.FileID > 0 && req.FolderID > 0) {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	if err := h.FileService.RenameByIDs(userID, req.FileID, req.FolderID, name); err != nil {
		if errors.Is(err, errors.ErrFileNotFound) || errors.Is(err, errors.ErrFolderNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在")
			return
		}
		if errors.Is(err, errors.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在或无权访问")
		return
	}
	response.Success(c, nil)
}

// MoveFileByID godoc
// @Summary 移动文件或文件夹
// @Description 通过 file_id 或 folder_id 指定要移动的对象，target_folder_id 指定目标目录（0 表示根目录）
// @Tags file
// @Accept json
// @Produce json
// @Param data body dto.MoveFileReq true "移动参数"
// @Success 200 {object} response.Response "成功返回（code=0）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/move [post]
func (h *FileHandler) MoveFileByID(c *gin.Context) {
	var req dto.MoveFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	if (req.FileID == 0 && req.FolderID == 0) || (req.FileID > 0 && req.FolderID > 0) {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	if err := h.FileService.MoveByIDs(userID, req.FileID, req.FolderID, req.TargetFolderID); err != nil {
		if errors.Is(err, errors.ErrFolderNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "目标目录不存在")
			return
		}
		if errors.Is(err, errors.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusBadRequest, response.CodeInvalidParam, "移动失败：目标目录无效")
		return
	}
	response.Success(c, nil)
}

// DeleteFileByID godoc
// @Summary 删除文件或文件夹
// @Description 通过 file_id 或 folder_id 删除对象（两者必须且只能传一个）
// @Tags file
// @Accept json
// @Produce json
// @Param data body dto.DeleteFileReq true "删除参数"
// @Success 200 {object} response.Response "成功返回（code=0）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/delete [post]
func (h *FileHandler) DeleteFileByID(c *gin.Context) {
	var req dto.DeleteFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	if (req.FileID == 0 && req.FolderID == 0) || (req.FileID > 0 && req.FolderID > 0) {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	if err := h.FileService.DeleteByIDs(userID, req.FileID, req.FolderID); err != nil {
		if errors.Is(err, errors.ErrFileNotFound) || errors.Is(err, errors.ErrFolderNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在")
			return
		}
		if errors.Is(err, errors.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "文件不存在或无权访问")
		return
	}
	response.Success(c, nil)
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
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
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

// GetPickUpCodeListByUserIDAndPage godoc
// @Summary 获取取件码列表
// @Description 分页获取当前用户的所有取件码列表
// @Tags file
// @Accept json
// @Produce json
// @Param page query int true "页码" default(1)
// @Param page_size query int true "每页数量" default(10)
// @Success 200 {object} response.Response{data=[]vo.PickUpCodeListItem} "成功返回（code=0,data=取件码列表）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/code/list [get]
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

// GetPickUpCodeListCountByUserID godoc
// @Summary 获取取件码总数
// @Description 获取当前用户的取件码总数
// @Tags file
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=int64} "成功返回（code=0,data=总数）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Security ApiKeyAuth
// @Router /file/code/count [get]
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

// DeletePickUpCodeByID godoc
// @Summary 删除取件码
// @Description 删除指定的取件码
// @Tags file
// @Accept json
// @Produce json
// @Param id query int true "取件码ID"
// @Success 200 {object} response.Response "成功返回（code=0）"
// @Failure 401 {object} map[string]string "未授权（缺少/无效 Authorization）"
// @Failure 403 {object} map[string]string "权限不足"
// @Failure 404 {object} map[string]string "取件码不存在"
// @Security ApiKeyAuth
// @Router /file/code [delete]
func (h *FileHandler) DeletePickUpCodeByID(c *gin.Context) {
	var req dto.DeletePickUpCodeReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	if err := h.FileService.DeletePickUpCodeByID(userID, req.ID); err != nil {
		if errors.Is(err, service.ErrPickupTargetNotFound) {
			response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "取件码不存在")
			return
		}
		if errors.Is(err, service.ErrPermissionDenied) {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeUnauthorized, "权限不足")
			return
		}
		response.FailWithStatus(c, http.StatusNotFound, response.CodeNotFound, "取件码不存在或无权访问")
		return
	}
	response.Success(c, nil)
}

// DownloadByPickUpCode godoc
// @Summary 通过取件码下载文件
// @Description 通过取件码下载文件或文件夹（无需登录）
// @Tags file
// @Accept json
// @Produce octet-stream
// @Param code query string true "取件码（6位字母数字组合）"
// @Success 200 {file} binary "文件内容流"
// @Failure 400 {object} map[string]string "取件码格式错误"
// @Failure 404 {object} map[string]string "取件码已失效或资源不存在"
// @Router /file/pickup/download [get]
func (h *FileHandler) DownloadByPickUpCode(c *gin.Context) {
	var req dto.DownloadByPickUpCodeReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	code, ok := normalizePickUpCode(req.Code)
	if !ok {
		response.Fail(c, response.CodeInvalidParam)
		return
	}

	setMeta := func(fileName, contentType string) {
		escapedName := url.PathEscape(fileName)
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+escapedName)
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Cache-Control", "no-cache")
	}

	if err := h.FileService.DownloadByPickUpCode(code, c.Writer, setMeta); err != nil {
		if errors.Is(err, service.ErrPickupCodeExpired) {
			response.FailWithMsg(c, response.CodeNotFound, "取件码已失效")
			return
		}
		if errors.Is(err, service.ErrPickupTargetNotFound) {
			response.FailWithMsg(c, response.CodeNotFound, "资源不存在")
			return
		}
		if errors.Is(err, service.ErrPickupEmptyFolder) {
			response.FailWithMsg(c, response.CodeNotFound, "文件夹为空")
			return
		}
		if errors.Is(err, os.ErrNotExist) || os.IsNotExist(err) {
			response.FailWithMsg(c, response.CodeNotFound, "文件不存在或已被删除")
			return
		}
		response.FailWithMsg(c, response.CodeServerError, err.Error())
		return
	}
}
