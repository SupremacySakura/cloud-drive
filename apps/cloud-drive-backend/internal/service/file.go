package service

import (
	"archive/zip"
	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/repository"
	"cloud-drive-backend/internal/utils"
	"cloud-drive-backend/internal/vo"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrPickupCodeExpired    = errors.New("pickup code expired")
	ErrPickupTargetNotFound = errors.New("pickup target not found")
	ErrPickupEmptyFolder    = errors.New("empty folder")
	ErrPublicShareNotFound  = errors.New("public share not found")
	ErrInvalidFileName      = errors.New("invalid file name")
	ErrStorageQuotaExceeded = errors.New("storage quota exceeded")
	ErrChunkSizeMismatch    = errors.New("chunk size mismatch")
	ErrInvalidMIMEType      = errors.New("invalid mime type")
	// 新增错误类型，用于HTTP状态码映射
	ErrFileNotFound     = errors.New("file not found")
	ErrFolderNotFound   = errors.New("folder not found")
	ErrPermissionDenied = errors.New("permission denied")
)

var allowedMIMETypes = map[string]bool{
	"image/":        true,
	"video/":        true,
	"application/pdf": true,
	"application/zip": true,
	"text/":         true,
}

func (s *fileService) IsAllowedMIMEType(mimeType string) bool {
	if mimeType == "" {
		return false
	}
	if allowedMIMETypes[mimeType] {
		return true
	}
	for prefix := range allowedMIMETypes {
		if strings.HasSuffix(prefix, "/") && strings.HasPrefix(mimeType, prefix) {
			return true
		}
	}
	return false
}

var reservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true,
	"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
	"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

func sanitizeFileName(name string) (string, error) {
	if name == "" {
		return "", ErrInvalidFileName
	}
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "\x00", "")
	var cleaned strings.Builder
	for _, r := range name {
		if r >= 32 && r != 127 {
			cleaned.WriteRune(r)
		}
	}
	name = cleaned.String()
	if name == "" || name == "." || name == ".." {
		return "", ErrInvalidFileName
	}
	upperName := strings.ToUpper(name)
	if reservedNames[upperName] {
		return "", ErrInvalidFileName
	}
	if ext := filepath.Ext(upperName); ext != "" {
		baseName := strings.TrimSuffix(upperName, ext)
		if reservedNames[baseName] {
			return "", ErrInvalidFileName
		}
	}
	dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range dangerousChars {
		if strings.Contains(name, char) {
			return "", ErrInvalidFileName
		}
	}
	return name, nil
}

func sanitizeStorageFileExt(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return ""
	}
	var cleaned strings.Builder
	for _, r := range ext {
		if r >= 32 && r != 127 {
			cleaned.WriteRune(r)
		}
	}
	ext = cleaned.String()
	if ext == "." {
		return ""
	}
	dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range dangerousChars {
		if strings.Contains(ext, char) {
			return ""
		}
	}
	return ext
}

func validateZipEntryPath(entryPath string) error {
	cleanPath := path.Clean(entryPath)
	if cleanPath == "." || cleanPath == "" {
		return errors.New("invalid zip entry path: empty path")
	}
	if strings.HasPrefix(cleanPath, "/") {
		return errors.New("invalid zip entry path: absolute path not allowed")
	}
	for _, segment := range strings.Split(cleanPath, "/") {
		if segment == ".." {
			return errors.New("invalid zip entry path: contains path traversal")
		}
	}
	return nil
}

const (
	defaultStorageLimitBytes    uint64 = 1024 * 1024 * 1024
	dashboardRecentActivitySize int    = 8
)

type FileServiceOptions struct {
	ChunkStoragePath string
	FileStoragePath  string
}

// FileService 定义文件服务的对外接口（用于 DI/替换实现）
type FileService interface {
    InitUploadFile(req *model.UploadTask) (task *model.UploadTask, err error)
    UploadFileChunkStream(userID uint, chunk *dto.UploadChunkReq, reader io.Reader, chunkSize int64) error
    IsAllowedMIMEType(mimeType string) bool
    MergeUploadedChunks(userID uint, taskID uint) error
    GetDashboardOverview(userID uint, storageLimit uint64) (*dto.DashboardOverviewResp, error)
    GetListByFolderIDAndUserID(folderID uint, userID uint, page, pageSize int) ([]dto.FileListItem, error)
    GetListCountByFolderIDAndUserID(folderID uint, userID uint) (int64, error)
    MakeDirectory(folderID uint, name string, userID uint) (uint, error)
    RenameByIDs(userID uint, fileID, folderID uint, name string) error
    MoveByIDs(userID uint, fileID, folderID, targetFolderID uint) error
    DeleteByIDs(userID uint, fileID, folderID uint) error
    CreatePickUpCode(code *model.PickUpCodeModel) (uint, error)
    GetPickUpCodeListByUserID(userID uint, page int, pageSize int) ([]vo.PickUpCodeListItem, error)
    GetPickUpCodeListCountByUserID(userID uint) (int64, error)
    DeletePickUpCodeByID(userID uint, codeID uint) error
    CreatePublicShareLink(fileID uint, userID uint) (string, error)
    GetPublicShareLink(fileID uint, userID uint) (string, error)
    DeletePublicShareLink(fileID uint, userID uint) error
    OpenPublicShare(token string, writer io.Writer, setMeta func(fileName, contentType string)) error
    PreviewFileByID(fileID uint, userID uint, writer io.Writer, setMeta func(fileName, contentType string)) error
    DownloadByIDs(userID uint, fileID, folderID uint, writer io.Writer, setMeta func(fileName, contentType string)) error
    DownloadByPickUpCode(code string, writer io.Writer, setMeta func(fileName, contentType string)) error
}

// FileService 为对外暴露的接口，保持与原有实现一致

type fileService struct {
	FileRepository *repository.FileRepository
	FileServiceOptions
}

type PickUpDownloadTarget struct {
	CodeID       uint
	Type         model.PickUpTargetType
	FilePath     string
	FolderID     uint
	DownloadName string
}

func (s *fileService) ensureStorageQuota(userID uint, additionalSize uint64) error {
	if additionalSize == 0 {
		return nil
	}
	storageUsed, err := s.FileRepository.GetStorageUsedByUserID(userID)
	if err != nil {
		return err
	}
	if storageUsed+additionalSize > defaultStorageLimitBytes {
		return ErrStorageQuotaExceeded
	}
	return nil
}

func NewFileService(fileRepository *repository.FileRepository, options FileServiceOptions) FileService {
    return &fileService{
        FileRepository:     fileRepository,
        FileServiceOptions: options,
    }
}

func (s *fileService) InitUploadFile(req *model.UploadTask) (task *model.UploadTask, err error) {
	exists, err := s.FileRepository.CheckFileExistsInFolder(req.FileHash, req.UserID, req.FolderID)
	if err != nil {
		return nil, err
	}
	if exists {
		return s.createInstantCompleteTask(req)
	}

	existingTask, err := s.FileRepository.GetUploadTaskByHashAndUserID(req.FileHash, req.UserID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		req.Status = model.UploadStatusUploading
		err = s.FileRepository.CreateUploadTask(req)
		if err != nil {
			return nil, err
		}
		err = os.MkdirAll(s.ChunkStoragePath+"/"+strconv.FormatUint(uint64(req.ID), 10), 0755)
		if err != nil {
			return nil, err
		}
		return req, nil
	}

	if existingTask.Status == model.UploadStatusCompleted {
		return s.createInstantTransferTask(req, existingTask)
	}

	return existingTask, nil
}

func (s *fileService) createInstantCompleteTask(req *model.UploadTask) (*model.UploadTask, error) {
	task := &model.UploadTask{
		FileHash:       req.FileHash,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		ChunkSize:      req.ChunkSize,
		TotalChunks:    req.TotalChunks,
		UploadedChunks: model.IntSlice{},
		FileType:       req.FileType,
		FolderID:       req.FolderID,
		UserID:         req.UserID,
		Status:         model.UploadStatusCompleted,
	}

	if err := s.FileRepository.CreateUploadTask(task); err != nil {
		return nil, err
	}

	exists, err := s.FileRepository.CheckFileExistsInFolder(req.FileHash, req.UserID, req.FolderID)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := s.ensureStorageQuota(req.UserID, req.FileSize); err != nil {
			return nil, err
		}
		fileModel := &model.FileModel{
			UserID:   req.UserID,
			FolderID: req.FolderID,
			Name:     req.FileName,
			Size:     req.FileSize,
			Type:     req.FileType,
			FileHash: req.FileHash,
		}
		if err := s.FileRepository.Create(fileModel); err != nil {
			return nil, err
		}
	}

	for i := 0; i < req.TotalChunks; i++ {
		task.UploadedChunks = append(task.UploadedChunks, i)
	}

	return task, nil
}

func (s *fileService) createInstantTransferTask(req *model.UploadTask, existingTask *model.UploadTask) (*model.UploadTask, error) {
	_ = existingTask
	task := &model.UploadTask{
		FileHash:       req.FileHash,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		ChunkSize:      req.ChunkSize,
		TotalChunks:    req.TotalChunks,
		UploadedChunks: model.IntSlice{},
		FileType:       req.FileType,
		FolderID:       req.FolderID,
		UserID:         req.UserID,
		Status:         model.UploadStatusCompleted,
	}

	for i := 0; i < req.TotalChunks; i++ {
		task.UploadedChunks = append(task.UploadedChunks, i)
	}

	if err := s.ensureStorageQuota(req.UserID, req.FileSize); err != nil {
		return nil, err
	}
	if err := s.FileRepository.CreateUploadTask(task); err != nil {
		return nil, err
	}

	fileModel := &model.FileModel{
		UserID:   req.UserID,
		FolderID: req.FolderID,
		Name:     req.FileName,
		Size:     req.FileSize,
		Type:     req.FileType,
		FileHash: req.FileHash,
	}
	if err := s.FileRepository.Create(fileModel); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *fileService) UploadFileChunkStream(userID uint, req *dto.UploadChunkReq, reader io.Reader, chunkSize int64) error {
	return s.FileRepository.DB.Transaction(func(tx *gorm.DB) error {
		var task model.UploadTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", req.TaskID, userID).
			First(&task).Error; err != nil {
			return err
		}

		expectedChunkSize := int64(task.ChunkSize)
		if req.ChunkIndex == task.TotalChunks-1 {
			expectedChunkSize = int64(task.FileSize) - int64(req.ChunkIndex)*int64(task.ChunkSize)
		}

		if chunkSize > expectedChunkSize {
			return ErrChunkSizeMismatch
		}

		chunkDir := s.ChunkStoragePath + "/" + strconv.FormatUint(uint64(task.ID), 10)
		if err := os.MkdirAll(chunkDir, 0755); err != nil {
			return err
		}

		chunkPath := chunkDir + "/" + strconv.Itoa(req.ChunkIndex)
		tmpPath := chunkPath + ".tmp"

		dst, err := os.Create(tmpPath)
		if err != nil {
			return err
		}

		written, err := io.Copy(dst, reader)
		if err != nil {
			dst.Close()
			return err
		}
		dst.Close()

		if written != chunkSize {
			os.Remove(tmpPath)
			return ErrChunkSizeMismatch
		}

		if err := os.Rename(tmpPath, chunkPath); err != nil {
			return err
		}

		ok, err := utils.VerifyFileSHA256(chunkPath, req.ChunkHash)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("chunk hash mismatch")
		}

		task.UploadedChunks = append(task.UploadedChunks, req.ChunkIndex)
		sort.Ints(task.UploadedChunks)

		return tx.Save(&task).Error
	})
}

func (s *fileService) MergeUploadedChunks(userID uint, taskID uint) error {
	task, err := s.FileRepository.GetUploadTaskByIDAndUserID(taskID, userID)
	if err != nil {
		return err
	}
	if task.Status == model.UploadStatusCompleted {
		return nil
	}
	if !utils.HasAllChunks(task.UploadedChunks, task.TotalChunks) {
		return errors.New("not all chunks uploaded")
	}
	if len(task.FileHash) < 4 {
		return errors.New("invalid file hash")
	}

	var totalChunkSize int64
	for i := 0; i < task.TotalChunks; i++ {
		chunkPath := s.ChunkStoragePath + "/" + strconv.FormatUint(uint64(task.ID), 10) + "/" + strconv.Itoa(i)
		info, err := os.Stat(chunkPath)
		if err != nil {
			return errors.New("chunk file not found")
		}
		totalChunkSize += info.Size()
	}
	if uint64(totalChunkSize) != task.FileSize {
		return ErrChunkSizeMismatch
	}

	if err := s.ensureStorageQuota(userID, task.FileSize); err != nil {
		return err
	}

	dirPath := s.FileStoragePath + "/" + task.FileHash[0:2] + "/" + task.FileHash[2:4]
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	filePath := dirPath + "/" + task.FileHash + sanitizeStorageFileExt(task.FileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	for i := 0; i < task.TotalChunks; i++ {
		chunkPath := s.ChunkStoragePath + "/" + strconv.FormatUint(uint64(task.ID), 10) + "/" + strconv.Itoa(i)
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return err
		}
		_, err = file.Write(chunkData)
		if err != nil {
			return err
		}
	}
	if err := file.Close(); err != nil {
		return err
	}
	ok, err := utils.VerifyFileSHA256(filePath, task.FileHash)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("merged file hash mismatch")
	}
    // 使用事务确保数据库操作原子性
    err = s.FileRepository.DB.Transaction(func(tx *gorm.DB) error {
        task.Status = model.UploadStatusCompleted
        if err := tx.Save(task).Error; err != nil {
            return err
        }
        fileModel := &model.FileModel{
            UserID:   task.UserID,
            FolderID: task.FolderID,
            Name:     task.FileName,
            Size:     task.FileSize,
            Type:     task.FileType,
            FileHash: task.FileHash,
        }
        if err := tx.Create(fileModel).Error; err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        return err
    }
	return nil
}

func (s *fileService) GetDashboardOverview(userID uint, storageLimit uint64) (*dto.DashboardOverviewResp, error) {
	if storageLimit == 0 {
		storageLimit = defaultStorageLimitBytes
	}

	storageUsed, err := s.FileRepository.GetStorageUsedByUserID(userID)
	if err != nil {
		return nil, err
	}

	stats, err := s.FileRepository.GetFileStatsByUserID(userID)
	if err != nil {
		return nil, err
	}
	statMap := make(map[string]dto.DashboardFileStatItem, len(stats))
	for _, item := range stats {
		statMap[item.Type] = item
	}
	orderedTypes := []string{"image", "video", "audio", "document", "other"}
	fileStats := make([]dto.DashboardFileStatItem, 0, len(orderedTypes))
	for _, fileType := range orderedTypes {
		stat, ok := statMap[fileType]
		if !ok {
			stat = dto.DashboardFileStatItem{
				Type:  fileType,
				Count: 0,
				Size:  0,
			}
		}
		fileStats = append(fileStats, stat)
	}

	recentActivities, err := s.FileRepository.GetRecentActivitiesByUserID(userID, dashboardRecentActivitySize)
	if err != nil {
		return nil, err
	}

	usedPercent := int(storageUsed * 100 / storageLimit)
	if usedPercent > 100 {
		usedPercent = 100
	}
	storageLeft := uint64(0)
	if storageLimit > storageUsed {
		storageLeft = storageLimit - storageUsed
	}

	return &dto.DashboardOverviewResp{
		StorageUsed:        storageUsed,
		StorageTotal:       storageLimit,
		StorageLeft:        storageLeft,
		StorageUsedPercent: usedPercent,
		FileStats:          fileStats,
		RecentActivities:   recentActivities,
	}, nil
}

func (s *fileService) GetListByFolderIDAndUserID(folderID uint, userID uint, page, pageSize int) ([]dto.FileListItem, error) {
	list, err := s.FileRepository.GetListByFolderIDAndUserID(folderID, userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *fileService) GetListCountByFolderIDAndUserID(folderID uint, userID uint) (int64, error) {
	count, err := s.FileRepository.GetListCountByFolderIDAndUserID(folderID, userID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *fileService) MakeDirectory(folderID uint, name string, userID uint) (uint, error) {
    cleanedName, err := utils.SanitizeFileName(name)
	if err != nil {
		return 0, err
	}
	id, err := s.FileRepository.MakeDirectory(folderID, cleanedName, userID)
	return id, err
}

func (s *fileService) RenameByIDs(userID uint, fileID, folderID uint, name string) error {
    cleanedName, err := utils.SanitizeFileName(name)
	if err != nil {
		return fmt.Errorf("文件名无效: %w", err)
	}
	if fileID > 0 && folderID > 0 {
		return errors.New("invalid rename target")
	}
	if fileID == 0 && folderID == 0 {
		return errors.New("missing rename target")
	}
	if fileID > 0 {
		err = s.FileRepository.RenameFileByIDAndUserID(fileID, userID, cleanedName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFileNotFound
			}
			return fmt.Errorf("重命名文件失败: %w", err)
		}
		return nil
	}
	err = s.FileRepository.RenameFolderByIDAndUserID(folderID, userID, cleanedName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFolderNotFound
		}
		return fmt.Errorf("重命名文件夹失败: %w", err)
	}
	return nil
}

func (s *fileService) MoveByIDs(userID uint, fileID, folderID, targetFolderID uint) error {
	if fileID > 0 && folderID > 0 {
		return errors.New("invalid move target")
	}
	if fileID == 0 && folderID == 0 {
		return errors.New("missing move target")
	}

	if targetFolderID > 0 {
		if _, err := s.FileRepository.GetFolderByFolderIDAndUserID(targetFolderID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFolderNotFound
			}
			return fmt.Errorf("移动失败: %w", err)
		}
	}

	if fileID > 0 {
		err := s.FileRepository.MoveFileByIDAndUserID(fileID, userID, targetFolderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFileNotFound
			}
			return fmt.Errorf("移动文件失败: %w", err)
		}
		return nil
	}

	sourceFolder, err := s.FileRepository.GetFolderByFolderIDAndUserID(folderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFolderNotFound
		}
		return fmt.Errorf("移动文件夹失败: %w", err)
	}
	if sourceFolder.ID == targetFolderID {
		return errors.New("cannot move folder into itself")
	}

	// 防止把文件夹移动到自己的子孙目录中，避免形成环。
	current := targetFolderID
	for current != 0 {
		if current == sourceFolder.ID {
			return errors.New("cannot move folder into child folder")
		}
		parent, err := s.FileRepository.GetFolderByFolderIDAndUserID(current, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFolderNotFound
			}
			return fmt.Errorf("移动文件夹失败: %w", err)
		}
		current = parent.ParentID
	}

	err = s.FileRepository.MoveFolderByIDAndUserID(folderID, userID, targetFolderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFolderNotFound
		}
		return fmt.Errorf("移动文件夹失败: %w", err)
	}
	return nil
}

func (s *fileService) DeleteByIDs(userID uint, fileID, folderID uint) error {
	if fileID > 0 && folderID > 0 {
		return errors.New("invalid delete target")
	}
	if fileID == 0 && folderID == 0 {
		return errors.New("missing delete target")
	}

	if fileID > 0 {
		err := s.FileRepository.DeleteFileByIDAndUserID(fileID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFileNotFound
			}
			return fmt.Errorf("删除文件失败: %w", err)
		}
		return nil
	}

	return s.FileRepository.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := &repository.FileRepository{DB: tx}

		rootFolder, err := txRepo.GetFolderByFolderIDAndUserID(folderID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFolderNotFound
			}
			return fmt.Errorf("删除文件夹失败: %w", err)
		}
		folderIDs := []uint{rootFolder.ID}
		queue := []uint{rootFolder.ID}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			children, err := txRepo.GetDirectChildFoldersByParentAndUserID(current, userID)
			if err != nil {
				return fmt.Errorf("删除文件夹失败: %w", err)
			}
			for _, child := range children {
				folderIDs = append(folderIDs, child.ID)
				queue = append(queue, child.ID)
			}
		}

		if err := txRepo.DeleteFilesByFolderIDsAndUserID(folderIDs, userID); err != nil {
			return fmt.Errorf("删除文件失败: %w", err)
		}
		return txRepo.DeleteFoldersByIDsAndUserID(folderIDs, userID)
	})
}

func (s *fileService) CreatePickUpCode(code *model.PickUpCodeModel) (uint, error) {
	id, err := s.FileRepository.CreatePickUpCode(code)
	return id, err
}

func (s *fileService) GetPickUpCodeListByUserID(userID uint, page, pageSize int) ([]vo.PickUpCodeListItem, error) {
    // 先获取分页的 pickup 码列表
    list, err := s.FileRepository.GetPickUpCodeListByUserIDAndPage(userID, page, pageSize)
    if err != nil {
        return nil, err
    }

    // 批量聚合名称：先收集需要的 FileID / FolderID
    var fileIDs []uint
    var folderIDs []uint
    for _, item := range list {
        if item.FileID != nil {
            fileIDs = append(fileIDs, *item.FileID)
        }
        if item.FolderID != nil {
            folderIDs = append(folderIDs, *item.FolderID)
        }
    }

    // 通过批量查询获得名称映射，避免 N+1 查询
    fileMap := make(map[uint]string)
    if len(fileIDs) > 0 {
        files, err := s.FileRepository.GetFilesByIDs(fileIDs)
        if err != nil {
            return nil, err
        }
        for _, f := range files {
            fileMap[f.ID] = f.Name
        }
    }
    folderMap := make(map[uint]string)
    if len(folderIDs) > 0 {
        folders, err := s.FileRepository.GetFoldersByIDs(folderIDs)
        if err != nil {
            return nil, err
        }
        for _, fd := range folders {
            folderMap[fd.ID] = fd.Name
        }
    }

    var voList []vo.PickUpCodeListItem
    for _, item := range list {
        var name string
        if item.FileID != nil {
            if n, ok := fileMap[*item.FileID]; ok {
                name = n
            }
        } else if item.FolderID != nil {
            if n, ok := folderMap[*item.FolderID]; ok {
                name = n
            }
        }
        voList = append(voList, vo.PickUpCodeListItem{
            ID:          item.ID,
            Code:        item.Code,
            FileID:      item.FileID,
            FolderID:    item.FolderID,
            Name:        name,
            Type:        item.Type,
            Download:    int(item.Download),
            MaxDownload: int(item.MaxDownload),
            ExpireTime:  item.ExpireTime,
            CreatedAt:   item.CreatedAt,
            Status:      item.Status,
        })
    }
    return voList, nil
}

func (s *fileService) GetPickUpCodeListCountByUserID(userID uint) (int64, error) {
	count, err := s.FileRepository.GetPickUpCodeListCountByUserID(userID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *fileService) DeletePickUpCodeByID(userID uint, codeID uint) error {
	err := s.FileRepository.DeletePickUpCodeByIDAndUserID(codeID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPickupTargetNotFound
		}
		return fmt.Errorf("删除取件码失败: %w", err)
	}
	return nil
}

func (s *fileService) CreatePublicShareLink(fileID uint, userID uint) (string, error) {
	_, err := s.FileRepository.GetFileByFileIDAndUserID(fileID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrFileNotFound
		}
		return "", fmt.Errorf("创建分享链接失败: %w", err)
	}
	existing, err := s.FileRepository.GetPublicShareLinkByFileIDAndUserID(fileID, userID)
	if err == nil && existing != nil {
		return existing.Token, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	token, err := generatePublicShareToken()
	if err != nil {
		return "", err
	}
	link := &model.PublicShareLinkModel{
		Token:  token,
		FileID: fileID,
		UserID: userID,
	}
	if err := s.FileRepository.CreatePublicShareLink(link); err != nil {
		return "", err
	}
	return token, nil
}

func (s *fileService) GetPublicShareLink(fileID uint, userID uint) (string, error) {
	_, err := s.FileRepository.GetFileByFileIDAndUserID(fileID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrFileNotFound
		}
		return "", fmt.Errorf("获取分享链接失败: %w", err)
	}
	link, err := s.FileRepository.GetPublicShareLinkByFileIDAndUserID(fileID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrPublicShareNotFound
		}
		return "", err
	}
	return link.Token, nil
}

func (s *fileService) DeletePublicShareLink(fileID uint, userID uint) error {
	_, err := s.FileRepository.GetFileByFileIDAndUserID(fileID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFileNotFound
		}
		return fmt.Errorf("删除分享链接失败: %w", err)
	}
	err = s.FileRepository.DeletePublicShareLinkByFileIDAndUserID(fileID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPublicShareNotFound
		}
		return err
	}
	return nil
}

func (s *fileService) OpenPublicShare(token string, writer io.Writer, setMeta func(fileName, contentType string)) error {
	link, err := s.FileRepository.GetPublicShareLinkByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPublicShareNotFound
		}
		return err
	}
	fileModel, err := s.FileRepository.GetFileByID(link.FileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPublicShareNotFound
		}
		return err
	}
	filePath, err := s.BuildFileAbsolutePath(fileModel)
	if err != nil {
		return err
	}
	contentType := mime.TypeByExtension(filepath.Ext(fileModel.Name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if setMeta != nil {
		setMeta(fileModel.Name, contentType)
	}
	return s.StreamSingleFile(filePath, writer)
}

func (s *fileService) PreviewFileByID(fileID uint, userID uint, writer io.Writer, setMeta func(fileName, contentType string)) error {
	fileModel, err := s.FileRepository.GetFileByFileIDAndUserID(fileID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFileNotFound
		}
		return fmt.Errorf("预览文件失败: %w", err)
	}
	filePath, err := s.BuildFileAbsolutePath(fileModel)
	if err != nil {
		return err
	}
	contentType := mime.TypeByExtension(filepath.Ext(fileModel.Name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if setMeta != nil {
		setMeta(fileModel.Name, contentType)
	}
	return s.StreamSingleFile(filePath, writer)
}

func (s *fileService) DownloadByIDs(userID uint, fileID, folderID uint, writer io.Writer, setMeta func(fileName, contentType string)) error {
	if fileID > 0 && folderID > 0 {
		return errors.New("invalid download target")
	}
	if fileID == 0 && folderID == 0 {
		return errors.New("missing download target")
	}

	if fileID > 0 {
		fileModel, err := s.FileRepository.GetFileByFileIDAndUserID(fileID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFileNotFound
			}
			return fmt.Errorf("下载文件失败: %w", err)
		}
		filePath, err := s.BuildFileAbsolutePath(fileModel)
		if err != nil {
			return err
		}
		contentType := mime.TypeByExtension(filepath.Ext(fileModel.Name))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		if setMeta != nil {
			setMeta(fileModel.Name, contentType)
		}
		return s.StreamSingleFile(filePath, writer)
	}

	folderModel, err := s.FileRepository.GetFolderByFolderIDAndUserID(folderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFolderNotFound
		}
		return fmt.Errorf("下载文件夹失败: %w", err)
	}
	if setMeta != nil {
		setMeta(folderModel.Name+".zip", "application/zip")
	}
	return s.StreamFolderAsZip(folderModel.ID, writer)
}

func generatePublicShareToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *fileService) DownloadByPickUpCode(code string, writer io.Writer, setMeta func(fileName, contentType string)) error {
	target, err := s.ResolveActivePickUpCode(code)
	if err != nil {
		return err
	}

	contentType := "application/octet-stream"
	switch target.Type {
	case model.PickUpTargetTypeFile:
		contentType = mime.TypeByExtension(filepath.Ext(target.DownloadName))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		if setMeta != nil {
			setMeta(target.DownloadName, contentType)
		}
		if err := s.StreamSingleFile(target.FilePath, writer); err != nil {
			return err
		}
	case model.PickUpTargetTypeFolder:
		contentType = "application/zip"
		if setMeta != nil {
			setMeta(target.DownloadName, contentType)
		}
		if err := s.StreamFolderAsZip(target.FolderID, writer); err != nil {
			return err
		}
	default:
		return errors.New("invalid pickup target type")
	}

	if err := s.MarkPickUpDownloadSuccess(target.CodeID); err != nil {
		return err
	}
	return nil
}

func (s *fileService) ResolveActivePickUpCode(code string) (*PickUpDownloadTarget, error) {
	pickupCode, err := s.FileRepository.GetPickUpCodeByCode(code)
	if err != nil {
		return nil, err
	}
	if pickupCode.Status != model.PickUpCodeStatusActive {
		return nil, ErrPickupCodeExpired
	}
	now := time.Now()
	if now.After(pickupCode.ExpireTime) || pickupCode.Download >= pickupCode.MaxDownload {
		return nil, ErrPickupCodeExpired
	}

	switch pickupCode.Type {
	case model.PickUpTargetTypeFile:
		if pickupCode.FileID == nil {
			return nil, ErrPickupTargetNotFound
		}
		fileModel, err := s.FileRepository.GetFileByID(*pickupCode.FileID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPickupTargetNotFound
			}
			return nil, err
		}
		filePath, err := s.BuildFileAbsolutePath(fileModel)
		if err != nil {
			return nil, err
		}
		return &PickUpDownloadTarget{
			CodeID:       pickupCode.ID,
			Type:         pickupCode.Type,
			FilePath:     filePath,
			DownloadName: fileModel.Name,
		}, nil
	case model.PickUpTargetTypeFolder:
		if pickupCode.FolderID == nil {
			return nil, ErrPickupTargetNotFound
		}
		folderModel, err := s.FileRepository.GetFolderByID(*pickupCode.FolderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPickupTargetNotFound
			}
			return nil, err
		}
		return &PickUpDownloadTarget{
			CodeID:       pickupCode.ID,
			Type:         pickupCode.Type,
			FolderID:     folderModel.ID,
			DownloadName: folderModel.Name + ".zip",
		}, nil
	default:
		return nil, errors.New("invalid pickup target type")
	}
}

func (s *fileService) BuildFileAbsolutePath(fileModel *model.FileModel) (string, error) {
	if len(fileModel.FileHash) < 4 {
		return "", errors.New("invalid file hash")
	}
	ext := sanitizeStorageFileExt(fileModel.Name)
	filePath := s.FileStoragePath + "/" + fileModel.FileHash[0:2] + "/" + fileModel.FileHash[2:4] + "/" + fileModel.FileHash + ext
	if _, err := os.Stat(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func (s *fileService) StreamSingleFile(filePath string, writer io.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	return err
}

func (s *fileService) StreamFolderAsZip(folderID uint, writer io.Writer) error {
	zipWriter := zip.NewWriter(writer)
	rootFolder, err := s.FileRepository.GetFolderByID(folderID)
	if err != nil {
		return err
	}
    rootFolderName, err := utils.SanitizeFileName(rootFolder.Name)
    if err != nil {
        return err
    }
    if err := utils.ValidateZipEntryPath(rootFolderName); err != nil {
        return err
    }
	if err := s.writeFolderToZip(zipWriter, folderID, rootFolderName); err != nil {
		_ = zipWriter.Close()
		return err
	}
	return zipWriter.Close()
}

func (s *fileService) writeFolderToZip(zipWriter *zip.Writer, folderID uint, zipPrefix string) error {
	folders, files, err := s.FileRepository.GetChildrenByFolderID(folderID)
	if err != nil {
		return err
	}
	if len(folders) == 0 && len(files) == 0 {
		return ErrPickupEmptyFolder
	}

	for _, fileModel := range files {
		filePath, err := s.BuildFileAbsolutePath(&fileModel)
		if err != nil {
			return err
		}
		src, err := os.Open(filePath)
		if err != nil {
			return err
		}
        cleanedName, err := utils.SanitizeFileName(fileModel.Name)
        if err != nil {
            src.Close()
            return err
        }
        entryPath := path.Join(zipPrefix, cleanedName)
        if err := utils.ValidateZipEntryPath(entryPath); err != nil {
            src.Close()
            return err
        }
		info, statErr := src.Stat()
		if statErr != nil {
			src.Close()
			return statErr
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			src.Close()
			return err
		}
		header.Name = entryPath
		header.Method = zip.Deflate
		dst, err := zipWriter.CreateHeader(header)
		if err != nil {
			src.Close()
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			return err
		}
		src.Close()
	}

        for _, folder := range folders {
        cleanedFolderName, err := utils.SanitizeFileName(folder.Name)
        if err != nil {
            return err
        }
		nextPrefix := path.Join(zipPrefix, cleanedFolderName)
		if err := validateZipEntryPath(nextPrefix); err != nil {
			return err
		}
		if err := s.writeFolderToZip(zipWriter, folder.ID, nextPrefix); err != nil {
			if errors.Is(err, ErrPickupEmptyFolder) {
				_, _ = zipWriter.Create(nextPrefix + "/")
				continue
			}
			return err
		}
	}
	return nil
}

func (s *fileService) MarkPickUpDownloadSuccess(codeID uint) error {
	err := s.FileRepository.IncrementDownloadAndMaybeExpire(codeID, time.Now())
	if err != nil && strings.Contains(err.Error(), "pickup code expired") {
		return ErrPickupCodeExpired
	}
	return err
}
