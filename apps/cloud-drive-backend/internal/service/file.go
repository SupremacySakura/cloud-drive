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
	ErrPickupCodeExpired     = errors.New("pickup code expired")
	ErrPickupTargetNotFound  = errors.New("pickup target not found")
	ErrPickupEmptyFolder     = errors.New("empty folder")
	ErrPublicShareNotFound   = errors.New("public share not found")
	ErrInvalidFileName       = errors.New("invalid file name")
	ErrStorageQuotaExceeded  = errors.New("storage quota exceeded")
	ErrChunkSizeMismatch     = errors.New("chunk size mismatch")
	ErrChunkHashMismatch     = errors.New("chunk hash mismatch")
	ErrChunkConflict         = errors.New("chunk content conflicts with an existing upload")
	ErrUploadStateConflict   = errors.New("upload task is not accepting chunks")
	ErrUploadIncomplete      = errors.New("not all chunks uploaded")
	ErrInvalidMIMEType       = errors.New("invalid mime type")
	ErrUploadRequestConflict = errors.New("upload request conflicts with an existing idempotency key")
	// 新增错误类型，用于HTTP状态码映射
	ErrFileNotFound            = errors.New("file not found")
	ErrFolderNotFound          = errors.New("folder not found")
	ErrPermissionDenied        = errors.New("permission denied")
	errUploadTaskAlreadyExists = errors.New("upload task already exists")
)

var allowedMIMETypes = map[string]bool{
	"image/":          true,
	"video/":          true,
	"application/pdf": true,
	"application/zip": true,
	"text/":           true,
}

var blockedMIMETypes = map[string]bool{
	"text/javascript":          true,
	"application/javascript":   true,
	"application/x-javascript": true,
	"application/ecmascript":   true,
	"text/ecmascript":          true,
}

func (s *fileService) IsAllowedMIMEType(mimeType string) bool {
	if mimeType == "" {
		return false
	}
	if blockedMIMETypes[mimeType] {
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

func sanitizeFileName(name string) (string, error) {
	normalizedName, err := utils.NormalizeUploadFileName(name)
	if err != nil {
		return "", ErrInvalidFileName
	}
	return normalizedName, nil
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

func buildUploadIdentity(task *model.UploadTask) (idempotencyKey, requestHash, normalizedName string, err error) {
	idempotencyKey, requestHash, normalizedName, err = utils.BuildUploadIdentity(task)
	if err != nil {
		return "", "", "", ErrInvalidFileName
	}
	return idempotencyKey, requestHash, normalizedName, nil
}

func containsChunk(chunks model.IntSlice, target int) bool {
	for _, chunk := range chunks {
		if chunk == target {
			return true
		}
	}
	return false
}

func withoutChunk(chunks model.IntSlice, target int) model.IntSlice {
	result := make(model.IntSlice, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk != target {
			result = append(result, chunk)
		}
	}
	return result
}

func validateZipEntryPath(entryPath string) error {
	if entryPath == "" {
		return errors.New("invalid zip entry path: empty path")
	}
	if strings.Contains(entryPath, "..") {
		return errors.New("invalid zip entry path: contains path traversal")
	}
	cleanPath := path.Clean(entryPath)
	if cleanPath == "." {
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
	MergeUploadedChunks(userID uint, taskID uint) (*vo.MergeUploadedChunksResp, error)
	GetDashboardOverview(userID uint, storageLimit uint64) (*dto.DashboardOverviewResp, error)
	GetListByFolderIDAndUserID(folderID uint, userID uint, page, pageSize int) ([]dto.FileListItem, error)
	GetListCountByFolderIDAndUserID(folderID uint, userID uint) (int64, error)
	MakeDirectory(folderID uint, name string, userID uint) (uint, error)
	RenameByIDs(userID uint, fileID, folderID uint, name string) error
	MoveByIDs(userID uint, fileID, folderID, targetFolderID uint) error
	DeleteByIDs(userID uint, fileID, folderID uint) error
	CreatePickUpCode(userID uint, code *model.PickUpCodeModel) (uint, error)
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
	newMergeLeaseID func() (string, error)
	now             func() time.Time
}

type PickUpDownloadTarget struct {
	CodeID       uint
	UserID       uint
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

func (s *fileService) buildStorageDir(fileHash string) (string, error) {
	if len(fileHash) < 4 {
		return "", errors.New("invalid file hash")
	}
	return filepath.Join(s.FileStoragePath, fileHash[0:2], fileHash[2:4]), nil
}

func (s *fileService) buildCanonicalStoragePath(fileHash string) (string, error) {
	dir, err := s.buildStorageDir(fileHash)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileHash), nil
}

func NewFileService(fileRepository *repository.FileRepository, options FileServiceOptions) FileService {
	return &fileService{
		FileRepository:     fileRepository,
		FileServiceOptions: options,
	}
}

func (s *fileService) nextMergeLeaseID() (string, error) {
	if s.newMergeLeaseID != nil {
		return s.newMergeLeaseID()
	}
	return generateRandomHex(16)
}

func (s *fileService) currentTime() time.Time {
	if s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

func (s *fileService) InitUploadFile(req *model.UploadTask) (task *model.UploadTask, err error) {
	idempotencyKey, requestHash, normalizedName, err := buildUploadIdentity(req)
	if err != nil {
		return nil, err
	}
	req.IdempotencyKey = idempotencyKey
	req.RequestHash = requestHash
	req.FileName = normalizedName
	req.FileHash = strings.ToLower(req.FileHash)

	existingTask, err := s.FileRepository.GetUploadTaskByIdempotencyKey(idempotencyKey)
	if err == nil {
		if existingTask.RequestHash != requestHash {
			return nil, ErrUploadRequestConflict
		}
		return existingTask, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.FolderID > 0 {
		if _, err := s.FileRepository.GetFolderByFolderIDAndUserID(req.FolderID, req.UserID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrFolderNotFound
			}
			return nil, err
		}
	}

	var instantFile *model.FileModel
	targetFile, err := s.FileRepository.GetFileByHashAndUserIDAndFolderID(req.FileHash, req.UserID, req.FolderID)
	switch {
	case err == nil:
		req.Status = model.UploadStatusCompleted
		req.FileID = &targetFile.ID
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	default:
		_, blobErr := s.FileRepository.GetFileByHash(req.FileHash)
		if blobErr == nil {
			if err := s.ensureStorageQuota(req.UserID, req.FileSize); err != nil {
				return nil, err
			}
			req.Status = model.UploadStatusCompleted
			instantFile = &model.FileModel{
				UserID: req.UserID, FolderID: req.FolderID, Name: req.FileName,
				Size: req.FileSize, Type: req.FileType, FileHash: req.FileHash,
			}
		} else if errors.Is(blobErr, gorm.ErrRecordNotFound) {
			req.Status = model.UploadStatusUploading
		} else {
			return nil, blobErr
		}
	}

	created := false
	err = s.FileRepository.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(req)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errUploadTaskAlreadyExists
		}
		created = true
		if instantFile == nil {
			return nil
		}
		if err := tx.Create(instantFile).Error; err != nil {
			return err
		}
		req.FileID = &instantFile.ID
		return tx.Model(req).Update("file_id", instantFile.ID).Error
	})
	if errors.Is(err, errUploadTaskAlreadyExists) {
		existingTask, getErr := s.FileRepository.GetUploadTaskByIdempotencyKey(idempotencyKey)
		if getErr != nil {
			return nil, getErr
		}
		if existingTask.RequestHash != requestHash {
			return nil, ErrUploadRequestConflict
		}
		return existingTask, nil
	}
	if err != nil {
		return nil, err
	}
	if created && req.Status == model.UploadStatusUploading {
		chunkDir := filepath.Join(s.ChunkStoragePath, strconv.FormatUint(uint64(req.ID), 10))
		if err := os.MkdirAll(chunkDir, 0755); err != nil {
			return nil, err
		}
	}
	return req, nil
}

func (s *fileService) UploadFileChunkStream(userID uint, req *dto.UploadChunkReq, reader io.Reader, chunkSize int64) error {
	return s.FileRepository.DB.Transaction(func(tx *gorm.DB) error {
		var task model.UploadTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", req.TaskID, userID).
			First(&task).Error; err != nil {
			return err
		}

		if req.ChunkIndex < 0 || req.ChunkIndex >= task.TotalChunks {
			return ErrChunkSizeMismatch
		}

		expectedChunkSize := int64(task.ChunkSize)
		if req.ChunkIndex == task.TotalChunks-1 {
			expectedChunkSize = int64(task.FileSize) - int64(req.ChunkIndex)*int64(task.ChunkSize)
		}

		if chunkSize != expectedChunkSize {
			return ErrChunkSizeMismatch
		}
		req.ChunkHash = strings.ToLower(strings.TrimSpace(req.ChunkHash))
		chunkDir := filepath.Join(s.ChunkStoragePath, strconv.FormatUint(uint64(task.ID), 10))
		chunkPath := filepath.Join(chunkDir, strconv.Itoa(req.ChunkIndex))

		var existingChunk model.UploadChunk
		err := tx.Where("task_id = ? AND chunk_index = ?", req.TaskID, req.ChunkIndex).
			First(&existingChunk).Error
		if err == nil {
			if existingChunk.ChunkHash != req.ChunkHash || existingChunk.Size != chunkSize {
				return ErrChunkConflict
			}
			if task.Status == model.UploadStatusCompleted {
				return nil
			}
			info, statErr := os.Stat(chunkPath)
			if statErr == nil && info.Size() == chunkSize {
				valid, verifyErr := utils.VerifyFileSHA256(chunkPath, req.ChunkHash)
				if verifyErr != nil {
					return verifyErr
				}
				if valid {
					if containsChunk(task.UploadedChunks, req.ChunkIndex) {
						return nil
					}
					if task.Status != model.UploadStatusUploading {
						return ErrUploadStateConflict
					}
					task.UploadedChunks = append(task.UploadedChunks, req.ChunkIndex)
					sort.Ints(task.UploadedChunks)
					return tx.Model(&task).Update("uploaded_chunks", task.UploadedChunks).Error
				}
			}
			if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
				return statErr
			}
			if task.Status != model.UploadStatusUploading {
				return ErrUploadStateConflict
			}
			if err := tx.Delete(&existingChunk).Error; err != nil {
				return err
			}
			task.UploadedChunks = withoutChunk(task.UploadedChunks, req.ChunkIndex)
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if task.Status != model.UploadStatusUploading {
			return ErrUploadStateConflict
		}

		if err := os.MkdirAll(chunkDir, 0755); err != nil {
			return err
		}

		tmpToken, err := generateRandomHex(8)
		if err != nil {
			return err
		}
		tmpPath := chunkPath + ".tmp-" + tmpToken
		defer os.Remove(tmpPath)

		dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
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

		ok, err := utils.VerifyFileSHA256(tmpPath, req.ChunkHash)
		if err != nil {
			return err
		}
		if !ok {
			return ErrChunkHashMismatch
		}
		if err := os.Rename(tmpPath, chunkPath); err != nil {
			return err
		}

		if !containsChunk(task.UploadedChunks, req.ChunkIndex) {
			task.UploadedChunks = append(task.UploadedChunks, req.ChunkIndex)
			sort.Ints(task.UploadedChunks)
		}

		chunkRecord := &model.UploadChunk{
			TaskID: req.TaskID, ChunkIndex: req.ChunkIndex,
			ChunkHash: req.ChunkHash, Size: chunkSize,
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(chunkRecord)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var winner model.UploadChunk
			if err := tx.Where("task_id = ? AND chunk_index = ?", req.TaskID, req.ChunkIndex).First(&winner).Error; err != nil {
				return err
			}
			if winner.ChunkHash != req.ChunkHash || winner.Size != chunkSize {
				return ErrChunkConflict
			}
		}

		return tx.Model(&task).Update("uploaded_chunks", task.UploadedChunks).Error
	})
}

func (s *fileService) MergeUploadedChunks(userID uint, taskID uint) (*vo.MergeUploadedChunksResp, error) {
	leaseID, err := s.nextMergeLeaseID()
	if err != nil {
		return nil, err
	}
	now := s.currentTime()
	leaseExpiresAt := now.Add(10 * time.Minute)
	acquire := s.FileRepository.DB.Model(&model.UploadTask{}).
		Where("id = ? AND user_id = ? AND (status = ? OR (status = ? AND (merge_lease_expires_at IS NULL OR merge_lease_expires_at < ?)))",
			taskID, userID, model.UploadStatusUploading, model.UploadStatusMerging, now).
		Updates(map[string]interface{}{
			"status":                 model.UploadStatusMerging,
			"merge_lease_id":         leaseID,
			"merge_lease_expires_at": leaseExpiresAt,
		})
	if acquire.Error != nil {
		return nil, acquire.Error
	}
	if acquire.RowsAffected == 0 {
		task, err := s.FileRepository.GetUploadTaskByIDAndUserID(taskID, userID)
		if err != nil {
			return nil, err
		}
		switch task.Status {
		case model.UploadStatusMerging:
			return &vo.MergeUploadedChunksResp{Status: model.UploadStatusMerging}, nil
		case model.UploadStatusCompleted:
			return &vo.MergeUploadedChunksResp{Status: model.UploadStatusCompleted, FileID: task.FileID}, nil
		default:
			return nil, ErrUploadStateConflict
		}
	}

	mergeCompleted := false
	defer func() {
		if mergeCompleted {
			return
		}
		_ = s.FileRepository.DB.Model(&model.UploadTask{}).
			Where("id = ? AND user_id = ? AND merge_lease_id = ?", taskID, userID, leaseID).
			Updates(map[string]interface{}{
				"status":                 model.UploadStatusUploading,
				"merge_lease_id":         "",
				"merge_lease_expires_at": nil,
			}).Error
	}()

	task, err := s.FileRepository.GetUploadTaskByIDAndUserID(taskID, userID)
	if err != nil {
		return nil, err
	}
	if task.MergeLeaseID != leaseID || task.Status != model.UploadStatusMerging {
		return nil, ErrUploadStateConflict
	}
	if !utils.HasAllChunks(task.UploadedChunks, task.TotalChunks) {
		return nil, ErrUploadIncomplete
	}
	if len(task.FileHash) < 4 {
		return nil, errors.New("invalid file hash")
	}

	var totalChunkSize int64
	chunkDir := filepath.Join(s.ChunkStoragePath, strconv.FormatUint(uint64(task.ID), 10))
	for i := 0; i < task.TotalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, strconv.Itoa(i))
		info, err := os.Stat(chunkPath)
		if err != nil {
			return nil, errors.New("chunk file not found")
		}
		totalChunkSize += info.Size()
	}
	if uint64(totalChunkSize) != task.FileSize {
		return nil, ErrChunkSizeMismatch
	}
	if err := s.ensureStorageQuota(userID, task.FileSize); err != nil {
		return nil, err
	}

	dirPath, err := s.buildStorageDir(task.FileHash)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}
	filePath, err := s.buildCanonicalStoragePath(task.FileHash)
	if err != nil {
		return nil, err
	}
	tmpPath := filePath + ".merge-" + leaseID
	defer os.Remove(tmpPath)
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	for i := 0; i < task.TotalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, strconv.Itoa(i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			_ = file.Close()
			return nil, err
		}
		_, copyErr := io.Copy(file, chunkFile)
		closeErr := chunkFile.Close()
		if copyErr != nil {
			_ = file.Close()
			return nil, copyErr
		}
		if closeErr != nil {
			_ = file.Close()
			return nil, closeErr
		}
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	ok, err := utils.VerifyFileSHA256(tmpPath, task.FileHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("merged file hash mismatch")
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		return nil, err
	}

	var fileID uint
	err = s.FileRepository.DB.Transaction(func(tx *gorm.DB) error {
		var lockedTask model.UploadTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", taskID, userID).First(&lockedTask).Error; err != nil {
			return err
		}
		if lockedTask.Status == model.UploadStatusCompleted && lockedTask.FileID != nil {
			fileID = *lockedTask.FileID
			return nil
		}
		if lockedTask.Status != model.UploadStatusMerging || lockedTask.MergeLeaseID != leaseID {
			return ErrUploadStateConflict
		}
		fileModel := &model.FileModel{
			UserID: lockedTask.UserID, FolderID: lockedTask.FolderID, Name: lockedTask.FileName,
			Size: lockedTask.FileSize, Type: lockedTask.FileType, FileHash: lockedTask.FileHash,
		}
		if err := tx.Create(fileModel).Error; err != nil {
			return err
		}
		fileID = fileModel.ID
		return tx.Model(&lockedTask).Updates(map[string]interface{}{
			"status":                 model.UploadStatusCompleted,
			"file_id":                fileID,
			"merge_lease_id":         "",
			"merge_lease_expires_at": nil,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	mergeCompleted = true
	return &vo.MergeUploadedChunksResp{Status: model.UploadStatusCompleted, FileID: &fileID}, nil
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
	if folderID > 0 {
		if _, err := s.FileRepository.GetFolderByFolderIDAndUserID(folderID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, ErrFolderNotFound
			}
			return 0, err
		}
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

func (s *fileService) CreatePickUpCode(userID uint, code *model.PickUpCodeModel) (uint, error) {
	if code.UserID != userID {
		return 0, ErrPermissionDenied
	}
	if code.MaxDownload == 0 || time.Now().After(code.ExpireTime) {
		return 0, ErrPickupCodeExpired
	}
	switch code.Type {
	case model.PickUpTargetTypeFile:
		if code.FileID == nil || code.FolderID != nil {
			return 0, ErrPickupTargetNotFound
		}
		if _, err := s.FileRepository.GetFileByFileIDAndUserID(*code.FileID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, ErrPickupTargetNotFound
			}
			return 0, err
		}
	case model.PickUpTargetTypeFolder:
		if code.FolderID == nil || code.FileID != nil {
			return 0, ErrPickupTargetNotFound
		}
		if _, err := s.FileRepository.GetFolderByFolderIDAndUserID(*code.FolderID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, ErrPickupTargetNotFound
			}
			return 0, err
		}
	default:
		return 0, ErrPickupTargetNotFound
	}
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
	return s.StreamFolderAsZip(userID, folderModel.ID, writer)
}

func generateRandomHex(byteCount int) (string, error) {
	buf := make([]byte, byteCount)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func generatePublicShareToken() (string, error) {
	return generateRandomHex(24)
}

func (s *fileService) DownloadByPickUpCode(code string, writer io.Writer, setMeta func(fileName, contentType string)) error {
	target, err := s.ResolveActivePickUpCode(code)
	if err != nil {
		return err
	}
	if err := s.MarkPickUpDownloadSuccess(target.CodeID); err != nil {
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
		if err := s.StreamFolderAsZip(target.UserID, target.FolderID, writer); err != nil {
			return err
		}
	default:
		return errors.New("invalid pickup target type")
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
		fileModel, err := s.FileRepository.GetFileByFileIDAndUserID(*pickupCode.FileID, pickupCode.UserID)
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
			UserID:       pickupCode.UserID,
			Type:         pickupCode.Type,
			FilePath:     filePath,
			DownloadName: fileModel.Name,
		}, nil
	case model.PickUpTargetTypeFolder:
		if pickupCode.FolderID == nil {
			return nil, ErrPickupTargetNotFound
		}
		folderModel, err := s.FileRepository.GetFolderByFolderIDAndUserID(*pickupCode.FolderID, pickupCode.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPickupTargetNotFound
			}
			return nil, err
		}
		return &PickUpDownloadTarget{
			CodeID:       pickupCode.ID,
			UserID:       pickupCode.UserID,
			Type:         pickupCode.Type,
			FolderID:     folderModel.ID,
			DownloadName: folderModel.Name + ".zip",
		}, nil
	default:
		return nil, errors.New("invalid pickup target type")
	}
}

func (s *fileService) BuildFileAbsolutePath(fileModel *model.FileModel) (string, error) {
	filePath, err := s.buildCanonicalStoragePath(fileModel.FileHash)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		dirPath, dirErr := s.buildStorageDir(fileModel.FileHash)
		if dirErr != nil {
			return "", dirErr
		}
		legacyMatches, globErr := filepath.Glob(filepath.Join(dirPath, fileModel.FileHash+".*"))
		if globErr != nil {
			return "", globErr
		}
		for _, match := range legacyMatches {
			if info, statErr := os.Stat(match); statErr == nil && !info.IsDir() {
				return match, nil
			}
		}
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

func (s *fileService) StreamFolderAsZip(userID uint, folderID uint, writer io.Writer) error {
	zipWriter := zip.NewWriter(writer)
	rootFolder, err := s.FileRepository.GetFolderByFolderIDAndUserID(folderID, userID)
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
	if err := s.writeFolderToZip(zipWriter, userID, folderID, rootFolderName); err != nil {
		_ = zipWriter.Close()
		return err
	}
	return zipWriter.Close()
}

func (s *fileService) writeFolderToZip(zipWriter *zip.Writer, userID uint, folderID uint, zipPrefix string) error {
	folders, files, err := s.FileRepository.GetChildrenByFolderIDAndUserID(folderID, userID)
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
		if err := s.writeFolderToZip(zipWriter, userID, folder.ID, nextPrefix); err != nil {
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
