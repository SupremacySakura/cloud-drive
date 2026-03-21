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
	// 判断任务是否已经存在
	task, err = s.FileRepository.GetUploadTaskByHashAndUserID(req.FileHash, req.UserID)
	if err != nil {
		// 数据库查询错误
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// 任务不存在，创建新任务
		req.Status = model.UploadStatusUploading
		err = s.FileRepository.CreateUploadTask(req)
		if err != nil {
			return nil, err
		}
		// 创建文件夹
		err = os.MkdirAll(s.ChunkStoragePath+"/"+strconv.FormatUint(uint64(req.ID), 10), 0755)
		if err != nil {
			return nil, err
		}
		return req, nil
	}
	// 任务已经存在
	return task, nil
}

func (s *fileService) UploadFileChunkStream(userID uint, req *dto.UploadChunkReq, reader io.Reader) error {
	return s.FileRepository.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查 task + 加锁
		var task model.UploadTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", req.TaskID, userID).
			First(&task).Error; err != nil {
			return err
		}

		// 2. 写文件（流式写入）
		chunkDir := s.ChunkStoragePath + "/" + strconv.FormatUint(uint64(task.ID), 10)
		if err := os.MkdirAll(chunkDir, 0755); err != nil {
			return err
		}

		chunkPath := chunkDir + "/" + strconv.Itoa(req.ChunkIndex)

		// 原子写入
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

		// 3. hash 校验（如果需要）
		ok, err := utils.VerifyFileSHA256(chunkPath, req.ChunkHash)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("chunk hash mismatch")
		}

		// 4. 更新 uploadedChunks
		task.UploadedChunks = append(task.UploadedChunks, req.ChunkIndex)
		sort.Ints(task.UploadedChunks)

		return tx.Save(&task).Error
	})
}

func (s *fileService) MergeUploadedChunks(userID uint, taskID uint) error {
	// 从数据库获取任务
	task, err := s.FileRepository.GetUploadTaskByIDAndUserID(taskID, userID)
	if err != nil {
		return err
	}
	if task.Status == model.UploadStatusCompleted {
		return nil
	}
	// 检查是否所有分片都已上传
	if !utils.HasAllChunks(task.UploadedChunks, task.TotalChunks) {
		return errors.New("not all chunks uploaded")
	}
	// 合并分片
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
	// 更新任务状态为已完成
	task.Status = model.UploadStatusCompleted
	err = s.FileRepository.UpdateUploadTask(task)
	if err != nil {
		return err
	}
	// 创建文件记录
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

