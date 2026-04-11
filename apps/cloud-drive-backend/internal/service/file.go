package service

import (
	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/repository"
	"cloud-drive-backend/internal/utils"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
}

type fileService struct {
	FileRepository *repository.FileRepository
	FileServiceOptions
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
