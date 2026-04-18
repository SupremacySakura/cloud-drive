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
)

type FileServiceOptions struct {
	ChunkStoragePath string
	FileStoragePath  string
}

type FileService interface {
	InitUploadFile(req *model.UploadTask) (task *model.UploadTask, err error)
	UploadFileChunkStream(userID uint, chunk *dto.UploadChunkReq, reader io.Reader) error
	MergeUploadedChunks(userID uint, taskID uint) error
	GetListByFolderIDAndUserID(folderID uint, userID uint, page, pageSize int) ([]dto.FileListItem, error)
	GetListCountByFolderIDAndUserID(folderID uint, userID uint) (int64, error)
	MakeDirectory(folderID uint, name string, userID uint) (uint, error)
	CreatePickUpCode(code *model.PickUpCodeModel) (uint, error)
	GetPickUpCodeListByUserID(userID uint, page int, pageSize int) ([]vo.PickUpCodeListItem, error)
	GetPickUpCodeListCountByUserID(userID uint) (int64, error)
	CreatePublicShareLink(fileID uint, userID uint) (string, error)
	GetPublicShareLink(fileID uint, userID uint) (string, error)
	DeletePublicShareLink(fileID uint, userID uint) error
	OpenPublicShare(token string, writer io.Writer, setMeta func(fileName, contentType string)) error
	PreviewFileByID(fileID uint, userID uint, writer io.Writer, setMeta func(fileName, contentType string)) error
	DownloadByPickUpCode(code string, writer io.Writer, setMeta func(fileName, contentType string)) error
}

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

func (s *fileService) UploadFileChunkStream(userID uint, req *dto.UploadChunkReq, reader io.Reader) error {
	return s.FileRepository.DB.Transaction(func(tx *gorm.DB) error {
		var task model.UploadTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", req.TaskID, userID).
			First(&task).Error; err != nil {
			return err
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

		if _, err := io.Copy(dst, reader); err != nil {
			dst.Close()
			return err
		}

		dst.Close()

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
	dirPath := s.FileStoragePath + "/" + task.FileHash[0:2] + "/" + task.FileHash[2:4]
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	filePath := dirPath + "/" + task.FileHash + filepath.Ext(task.FileName)
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
	task.Status = model.UploadStatusCompleted
	err = s.FileRepository.UpdateUploadTask(task)
	if err != nil {
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
	err = s.FileRepository.Create(fileModel)
	if err != nil {
		return err
	}
	return nil
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
	id, err := s.FileRepository.MakeDirectory(folderID, name, userID)
	return id, err
}

func (s *fileService) CreatePickUpCode(code *model.PickUpCodeModel) (uint, error) {
	id, err := s.FileRepository.CreatePickUpCode(code)
	return id, err
}

func (s *fileService) GetPickUpCodeListByUserID(userID uint, page, pageSize int) ([]vo.PickUpCodeListItem, error) {
	list, err := s.FileRepository.GetPickUpCodeListByUserIDAndPage(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	var voList []vo.PickUpCodeListItem
	for _, item := range list {
		var name string
		if item.Type == model.PickUpTargetTypeFile {
			file, err := s.FileRepository.GetFileByFileIDAndUserID(*item.FileID, userID)
			if err != nil {
				continue
			}
			name = file.Name
		} else {
			folder, err := s.FileRepository.GetFolderByFolderIDAndUserID(*item.FolderID, userID)
			if err != nil {
				continue
			}
			name = folder.Name
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

func (s *fileService) CreatePublicShareLink(fileID uint, userID uint) (string, error) {
	_, err := s.FileRepository.GetFileByFileIDAndUserID(fileID, userID)
	if err != nil {
		return "", err
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
		return "", err
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
		return err
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
	filePath := s.FileStoragePath + "/" + fileModel.FileHash[0:2] + "/" + fileModel.FileHash[2:4] + "/" + fileModel.FileHash + filepath.Ext(fileModel.Name)
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
	if err := s.writeFolderToZip(zipWriter, folderID, rootFolder.Name); err != nil {
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
		entryPath := path.Join(zipPrefix, fileModel.Name)
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
		nextPrefix := path.Join(zipPrefix, folder.Name)
		if err := s.writeFolderToZip(zipWriter, folder.ID, nextPrefix); err != nil {
			if errors.Is(err, ErrPickupEmptyFolder) {
				_, _ = zipWriter.Create(path.Join(nextPrefix, "/"))
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
