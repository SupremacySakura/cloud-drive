package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/vo"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock FileService
type mockFileService struct {
	initUploadFileFunc          func(req *model.UploadTask) (*model.UploadTask, error)
	uploadFileChunkStreamFunc   func(userID uint, chunk *dto.UploadChunkReq, reader io.Reader, chunkSize int64) error
	isAllowedMIMETypeFunc       func(mimeType string) bool
	mergeUploadedChunksFunc     func(userID uint, taskID uint) error
	getDashboardOverviewFunc    func(userID uint, storageLimit uint64) (*dto.DashboardOverviewResp, error)
	getListByFolderIDFunc       func(folderID uint, userID uint, page, pageSize int) ([]dto.FileListItem, error)
	getListCountByFolderIDFunc  func(folderID uint, userID uint) (int64, error)
	makeDirectoryFunc           func(folderID uint, name string, userID uint) (uint, error)
	renameByIDsFunc             func(userID uint, fileID, folderID uint, name string) error
	moveByIDsFunc               func(userID uint, fileID, folderID, targetFolderID uint) error
	deleteByIDsFunc             func(userID uint, fileID, folderID uint) error
	createPickUpCodeFunc        func(code *model.PickUpCodeModel) (uint, error)
	getPickUpCodeListFunc       func(userID uint, page, pageSize int) ([]vo.PickUpCodeListItem, error)
	getPickUpCodeCountFunc      func(userID uint) (int64, error)
	deletePickUpCodeFunc        func(userID uint, codeID uint) error
	createPublicShareLinkFunc   func(fileID uint, userID uint) (string, error)
	getPublicShareLinkFunc      func(fileID uint, userID uint) (string, error)
	deletePublicShareLinkFunc   func(fileID uint, userID uint) error
	openPublicShareFunc         func(token string, writer io.Writer, setMeta func(fileName, contentType string)) error
	previewFileByIDFunc         func(fileID uint, userID uint, writer io.Writer, setMeta func(fileName, contentType string)) error
	downloadByIDsFunc           func(userID uint, fileID, folderID uint, writer io.Writer, setMeta func(fileName, contentType string)) error
	downloadByPickUpCodeFunc    func(code string, writer io.Writer, setMeta func(fileName, contentType string)) error
}

func (m *mockFileService) InitUploadFile(req *model.UploadTask) (*model.UploadTask, error) {
	if m.initUploadFileFunc != nil {
		return m.initUploadFileFunc(req)
	}
	return nil, nil
}

func (m *mockFileService) UploadFileChunkStream(userID uint, chunk *dto.UploadChunkReq, reader io.Reader, chunkSize int64) error {
	if m.uploadFileChunkStreamFunc != nil {
		return m.uploadFileChunkStreamFunc(userID, chunk, reader, chunkSize)
	}
	return nil
}

func (m *mockFileService) IsAllowedMIMEType(mimeType string) bool {
	if m.isAllowedMIMETypeFunc != nil {
		return m.isAllowedMIMETypeFunc(mimeType)
	}
	return true
}

func (m *mockFileService) MergeUploadedChunks(userID uint, taskID uint) error {
	if m.mergeUploadedChunksFunc != nil {
		return m.mergeUploadedChunksFunc(userID, taskID)
	}
	return nil
}

func (m *mockFileService) GetDashboardOverview(userID uint, storageLimit uint64) (*dto.DashboardOverviewResp, error) {
	if m.getDashboardOverviewFunc != nil {
		return m.getDashboardOverviewFunc(userID, storageLimit)
	}
	return nil, nil
}

func (m *mockFileService) GetListByFolderIDAndUserID(folderID uint, userID uint, page, pageSize int) ([]dto.FileListItem, error) {
	if m.getListByFolderIDFunc != nil {
		return m.getListByFolderIDFunc(folderID, userID, page, pageSize)
	}
	return nil, nil
}

func (m *mockFileService) GetListCountByFolderIDAndUserID(folderID uint, userID uint) (int64, error) {
	if m.getListCountByFolderIDFunc != nil {
		return m.getListCountByFolderIDFunc(folderID, userID)
	}
	return 0, nil
}

func (m *mockFileService) MakeDirectory(folderID uint, name string, userID uint) (uint, error) {
	if m.makeDirectoryFunc != nil {
		return m.makeDirectoryFunc(folderID, name, userID)
	}
	return 0, nil
}

func (m *mockFileService) RenameByIDs(userID uint, fileID, folderID uint, name string) error {
	if m.renameByIDsFunc != nil {
		return m.renameByIDsFunc(userID, fileID, folderID, name)
	}
	return nil
}

func (m *mockFileService) MoveByIDs(userID uint, fileID, folderID, targetFolderID uint) error {
	if m.moveByIDsFunc != nil {
		return m.moveByIDsFunc(userID, fileID, folderID, targetFolderID)
	}
	return nil
}

func (m *mockFileService) DeleteByIDs(userID uint, fileID, folderID uint) error {
	if m.deleteByIDsFunc != nil {
		return m.deleteByIDsFunc(userID, fileID, folderID)
	}
	return nil
}

func (m *mockFileService) CreatePickUpCode(code *model.PickUpCodeModel) (uint, error) {
	if m.createPickUpCodeFunc != nil {
		return m.createPickUpCodeFunc(code)
	}
	return 0, nil
}

func (m *mockFileService) GetPickUpCodeListByUserID(userID uint, page, pageSize int) ([]vo.PickUpCodeListItem, error) {
	if m.getPickUpCodeListFunc != nil {
		return m.getPickUpCodeListFunc(userID, page, pageSize)
	}
	return nil, nil
}

func (m *mockFileService) GetPickUpCodeListCountByUserID(userID uint) (int64, error) {
	if m.getPickUpCodeCountFunc != nil {
		return m.getPickUpCodeCountFunc(userID)
	}
	return 0, nil
}

func (m *mockFileService) DeletePickUpCodeByID(userID uint, codeID uint) error {
	if m.deletePickUpCodeFunc != nil {
		return m.deletePickUpCodeFunc(userID, codeID)
	}
	return nil
}

func (m *mockFileService) CreatePublicShareLink(fileID uint, userID uint) (string, error) {
	if m.createPublicShareLinkFunc != nil {
		return m.createPublicShareLinkFunc(fileID, userID)
	}
	return "", nil
}

func (m *mockFileService) GetPublicShareLink(fileID uint, userID uint) (string, error) {
	if m.getPublicShareLinkFunc != nil {
		return m.getPublicShareLinkFunc(fileID, userID)
	}
	return "", nil
}

func (m *mockFileService) DeletePublicShareLink(fileID uint, userID uint) error {
	if m.deletePublicShareLinkFunc != nil {
		return m.deletePublicShareLinkFunc(fileID, userID)
	}
	return nil
}

func (m *mockFileService) OpenPublicShare(token string, writer io.Writer, setMeta func(fileName, contentType string)) error {
	if m.openPublicShareFunc != nil {
		return m.openPublicShareFunc(token, writer, setMeta)
	}
	return nil
}

func (m *mockFileService) PreviewFileByID(fileID uint, userID uint, writer io.Writer, setMeta func(fileName, contentType string)) error {
	if m.previewFileByIDFunc != nil {
		return m.previewFileByIDFunc(fileID, userID, writer, setMeta)
	}
	return nil
}

func (m *mockFileService) DownloadByIDs(userID uint, fileID, folderID uint, writer io.Writer, setMeta func(fileName, contentType string)) error {
	if m.downloadByIDsFunc != nil {
		return m.downloadByIDsFunc(userID, fileID, folderID, writer, setMeta)
	}
	return nil
}

func (m *mockFileService) DownloadByPickUpCode(code string, writer io.Writer, setMeta func(fileName, contentType string)) error {
	if m.downloadByPickUpCodeFunc != nil {
		return m.downloadByPickUpCodeFunc(code, writer, setMeta)
	}
	return nil
}

func setupFileHandlerTest() (*gin.Engine, *mockFileService, *mockAuthService, *FileHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockFileSvc := &mockFileService{}
	mockAuthSvc := &mockAuthService{}
	handler := NewFileHandler(mockFileSvc, mockAuthSvc)
	return router, mockFileSvc, mockAuthSvc, handler
}

func setUserIDMiddleware(userID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// Test InitUploadFile Handler - 参数绑定验证
func TestInitUploadFile_InvalidJSON(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/init", setUserIDMiddleware(1), handler.InitUploadFile)

	req := httptest.NewRequest(http.MethodPost, "/init", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
}

func TestInitUploadFile_MissingRequiredFields(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/init", setUserIDMiddleware(1), handler.InitUploadFile)

	// 缺少必填字段
	reqBody, _ := json.Marshal(map[string]interface{}{
		"file_name": "test.txt",
		// 缺少 file_size, file_hash, chunk_size, total_chunks, file_type
	})
	req := httptest.NewRequest(http.MethodPost, "/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInitUploadFile_InvalidFileHash(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/init", setUserIDMiddleware(1), handler.InitUploadFile)

	reqBody, _ := json.Marshal(dto.InitUploadFileReq{
		FileName:    "test.txt",
		FileSize:    1024,
		FileHash:    "../etc/passwd", // 包含路径遍历
		ChunkSize:   1024,
		TotalChunks: 1,
		FileType:    "text/plain",
	})
	req := httptest.NewRequest(http.MethodPost, "/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "无效的文件哈希")
}

func TestInitUploadFile_FileTooLarge(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/init", setUserIDMiddleware(1), handler.InitUploadFile)

	reqBody, _ := json.Marshal(dto.InitUploadFileReq{
		FileName:    "test.txt",
		FileSize:    200 * 1024 * 1024, // 200MB，超过限制
		FileHash:    "abc123",
		ChunkSize:   1024,
		TotalChunks: 1,
		FileType:    "text/plain",
	})
	req := httptest.NewRequest(http.MethodPost, "/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "文件大小超过限制")
}

// Test GetListByFolderIDAndUserID Handler - Query参数绑定
func TestGetListByFolderIDAndUserID_InvalidQuery(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.GET("/list", setUserIDMiddleware(1), handler.GetListByFolderIDAndUserID)

	// 无效的分页参数
	req := httptest.NewRequest(http.MethodGet, "/list?folder_id=abc&page=-1&page_size=0", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetListByFolderIDAndUserID_ValidQuery(t *testing.T) {
	router, mockFileSvc, _, handler := setupFileHandlerTest()
	mockFileSvc.getListByFolderIDFunc = func(folderID uint, userID uint, page, pageSize int) ([]dto.FileListItem, error) {
		return []dto.FileListItem{}, nil
	}
	router.GET("/list", setUserIDMiddleware(1), handler.GetListByFolderIDAndUserID)

	req := httptest.NewRequest(http.MethodGet, "/list?folder_id=1&page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "0") // CodeSuccess
}

// Test MakeDirectory Handler
func TestMakeDirectory_InvalidJSON(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/mkdir", setUserIDMiddleware(1), handler.MakeDirectory)

	req := httptest.NewRequest(http.MethodPost, "/mkdir", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
}

func TestMakeDirectory_ValidRequest(t *testing.T) {
	router, mockFileSvc, _, handler := setupFileHandlerTest()
	mockFileSvc.makeDirectoryFunc = func(folderID uint, name string, userID uint) (uint, error) {
		return 123, nil
	}
	router.POST("/mkdir", setUserIDMiddleware(1), handler.MakeDirectory)

	reqBody, _ := json.Marshal(dto.MakeDirectoryReq{
		FolderID: 0,
		Name:     "newfolder",
	})
	req := httptest.NewRequest(http.MethodPost, "/mkdir", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "0") // CodeSuccess
}

// Test RenameFileByID Handler
func TestRenameFileByID_InvalidParams(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/rename", setUserIDMiddleware(1), handler.RenameFileByID)

	tests := []struct {
		name   string
		body   interface{}
		expect string
	}{
		{
			name: "无效JSON",
			body: "invalid",
			expect: "1001",
		},
		{
			name: "空名称",
			body: dto.RenameFileReq{FileID: 1, Name: "   "},
			expect: "1001",
		},
		{
			name: "同时指定file_id和folder_id",
			body: dto.RenameFileReq{FileID: 1, FolderID: 2, Name: "newname"},
			expect: "1001",
		},
		{
			name: "未指定file_id和folder_id",
			body: dto.RenameFileReq{Name: "newname"},
			expect: "1001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if str, ok := tt.body.(string); ok && str == "invalid" {
				reqBody = []byte("invalid json")
			} else {
				reqBody, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/rename", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), tt.expect)
		})
	}
}

// Test MoveFileByID Handler
func TestMoveFileByID_InvalidParams(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/move", setUserIDMiddleware(1), handler.MoveFileByID)

	tests := []struct {
		name string
		body dto.MoveFileReq
	}{
		{
			name: "同时指定file_id和folder_id",
			body: dto.MoveFileReq{FileID: 1, FolderID: 2, TargetFolderID: 3},
		},
		{
			name: "未指定file_id和folder_id",
			body: dto.MoveFileReq{TargetFolderID: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/move", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
		})
	}
}

// Test DeleteFileByID Handler
func TestDeleteFileByID_InvalidParams(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/delete", setUserIDMiddleware(1), handler.DeleteFileByID)

	tests := []struct {
		name string
		body dto.DeleteFileReq
	}{
		{
			name: "同时指定file_id和folder_id",
			body: dto.DeleteFileReq{FileID: 1, FolderID: 2},
		},
		{
			name: "未指定file_id和folder_id",
			body: dto.DeleteFileReq{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
		})
	}
}

// Test DownloadFileByID Handler
func TestDownloadFileByID_InvalidParams(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.GET("/download", setUserIDMiddleware(1), handler.DownloadFileByID)

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "未指定file_id和folder_id",
			query: "",
		},
		{
			name:  "同时指定file_id和folder_id",
			query: "file_id=1&folder_id=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/download?"+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
		})
	}
}

// Test CreatePublicShareLink Handler
func TestCreatePublicShareLink_InvalidJSON(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/share/link", setUserIDMiddleware(1), handler.CreatePublicShareLink)

	req := httptest.NewRequest(http.MethodPost, "/share/link", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
}

func TestCreatePublicShareLink_MissingFileID(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/share/link", setUserIDMiddleware(1), handler.CreatePublicShareLink)

	reqBody, _ := json.Marshal(dto.CreatePublicShareLinkReq{
		FileID: 0, // 无效的文件ID
	})
	req := httptest.NewRequest(http.MethodPost, "/share/link", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test normalizePickUpCode
func TestNormalizePickUpCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"ABC123", "ABC123", true},
		{"abc-123", "ABC123", true},
		{"  abc123  ", "ABC123", true},
		{"ABCD", "", false},     // 太短
		{"ABC1234", "", false},  // 太长
		{"ABC!23", "", false},   // 非法字符
		{"", "", false},         // 空字符串
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := normalizePickUpCode(tt.input)
			assert.Equal(t, tt.valid, ok)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test DeletePickUpCodeByID Handler
func TestDeletePickUpCodeByID_MissingID(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.DELETE("/code", setUserIDMiddleware(1), handler.DeletePickUpCodeByID)

	req := httptest.NewRequest(http.MethodDelete, "/code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeletePickUpCodeByID_InvalidID(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.DELETE("/code", setUserIDMiddleware(1), handler.DeletePickUpCodeByID)

	req := httptest.NewRequest(http.MethodDelete, "/code?id=abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test GetPublicShareLink Handler
func TestGetPublicShareLink_MissingFileID(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.GET("/share/link", setUserIDMiddleware(1), handler.GetPublicShareLink)

	req := httptest.NewRequest(http.MethodGet, "/share/link", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test OpenPublicShare Handler (无需认证)
func TestOpenPublicShare_MissingToken(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.GET("/share/open", handler.OpenPublicShare)

	req := httptest.NewRequest(http.MethodGet, "/share/open", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
}

func TestOpenPublicShare_ValidToken(t *testing.T) {
	router, mockFileSvc, _, handler := setupFileHandlerTest()
	mockFileSvc.openPublicShareFunc = func(token string, writer io.Writer, setMeta func(fileName, contentType string)) error {
		return nil
	}
	router.GET("/share/open", handler.OpenPublicShare)

	req := httptest.NewRequest(http.MethodGet, "/share/open?token=validtoken123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test DownloadByPickUpCode Handler (无需认证)
func TestDownloadByPickUpCode_InvalidCode(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.GET("/pickup/download", handler.DownloadByPickUpCode)

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "缺少code",
			query: "",
		},
		{
			name:  "无效code格式",
			query: "code=INVALID!",
		},
		{
			name:  "code长度不对",
			query: "code=ABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// URL编码
			params, _ := url.ParseQuery(tt.query)
			req := httptest.NewRequest(http.MethodGet, "/pickup/download?"+params.Encode(), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
		})
	}
}

// Test GetDashboardOverview Handler
func TestGetDashboardOverview_NoAuth(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.GET("/dashboard/overview", handler.GetDashboardOverview)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/overview", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1002") // CodeUnauthorized
}

// Test MergeUploadedChunks Handler
func TestMergeUploadedChunks_InvalidJSON(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/merge", setUserIDMiddleware(1), handler.MergeUploadedChunks)

	req := httptest.NewRequest(http.MethodPost, "/merge", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
}

func TestMergeUploadedChunks_ValidRequest(t *testing.T) {
	router, mockFileSvc, _, handler := setupFileHandlerTest()
	mockFileSvc.mergeUploadedChunksFunc = func(userID uint, taskID uint) error {
		return nil
	}
	router.POST("/merge", setUserIDMiddleware(1), handler.MergeUploadedChunks)

	reqBody, _ := json.Marshal(dto.MergeUploadedChunksReq{
		TaskID: 123,
	})
	req := httptest.NewRequest(http.MethodPost, "/merge", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "0") // CodeSuccess
}

// Test GetListCountByFolderIDAndUserID Handler
func TestGetListCountByFolderIDAndUserID_ValidQuery(t *testing.T) {
	router, mockFileSvc, _, handler := setupFileHandlerTest()
	mockFileSvc.getListCountByFolderIDFunc = func(folderID uint, userID uint) (int64, error) {
		return 10, nil
	}
	router.GET("/list/count", setUserIDMiddleware(1), handler.GetListCountByFolderIDAndUserID)

	req := httptest.NewRequest(http.MethodGet, "/list/count?folder_id=1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "0") // CodeSuccess
}

// Test PreviewFileByID Handler
func TestPreviewFileByID_MissingFileID(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.GET("/preview", setUserIDMiddleware(1), handler.PreviewFileByID)

	req := httptest.NewRequest(http.MethodGet, "/preview", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test DeletePublicShareLink Handler
func TestDeletePublicShareLink_MissingFileID(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.DELETE("/share/link", setUserIDMiddleware(1), handler.DeletePublicShareLink)

	req := httptest.NewRequest(http.MethodDelete, "/share/link", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test CreatePickUpCode Handler
func TestCreatePickUpCode_InvalidJSON(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/code", setUserIDMiddleware(1), handler.CreatePickUpCode)

	req := httptest.NewRequest(http.MethodPost, "/code", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
}

func TestCreatePickUpCode_InvalidCode(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.POST("/code", setUserIDMiddleware(1), handler.CreatePickUpCode)

	// 无效的取件码
	reqBody, _ := json.Marshal(dto.CreatePickUpCodeReq{
		Code:   "INVALID!",
		FileID: func() *uint { id := uint(1); return &id }(),
		Type:   model.PickUpTargetTypeFile,
	})
	req := httptest.NewRequest(http.MethodPost, "/code", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1001") // CodeInvalidParam
}

func TestCreatePickUpCode_ValidCode(t *testing.T) {
	router, mockFileSvc, _, handler := setupFileHandlerTest()
	mockFileSvc.createPickUpCodeFunc = func(code *model.PickUpCodeModel) (uint, error) {
		return 1, nil
	}
	router.POST("/code", setUserIDMiddleware(1), handler.CreatePickUpCode)

	reqBody, _ := json.Marshal(dto.CreatePickUpCodeReq{
		Code:   "ABC123",
		FileID: func() *uint { id := uint(1); return &id }(),
		Type:   model.PickUpTargetTypeFile,
	})
	req := httptest.NewRequest(http.MethodPost, "/code", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "0") // CodeSuccess
}

// Test GetPickUpCodeListByUserIDAndPage Handler
func TestGetPickUpCodeListByUserIDAndPage_InvalidPage(t *testing.T) {
	router, _, _, handler := setupFileHandlerTest()
	router.GET("/code/list", setUserIDMiddleware(1), handler.GetPickUpCodeListByUserIDAndPage)

	req := httptest.NewRequest(http.MethodGet, "/code/list?page=-1&page_size=0", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test InitUploadFile with User Not Found
func TestInitUploadFile_UserNotFound(t *testing.T) {
	router, _, mockAuthSvc, handler := setupFileHandlerTest()
	mockAuthSvc.getUserByIDFunc = func(userID uint) (*model.UserModel, error) {
		return nil, errors.New("user not found")
	}
	router.POST("/init", setUserIDMiddleware(1), handler.InitUploadFile)

	reqBody, _ := json.Marshal(dto.InitUploadFileReq{
		FileName:    "test.txt",
		FileSize:    1024,
		FileHash:    "abc123",
		ChunkSize:   1024,
		TotalChunks: 1,
		FileType:    "text/plain",
	})
	req := httptest.NewRequest(http.MethodPost, "/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1004") // CodeServerError
}
