package service

import (
	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/repository"
)

type FileService interface {
	InitUploadFile(req *model.UploadTask) (task *model.UploadTask, err error)
	UploadFileChunk() error
}

type fileService struct {
	FileRepository *repository.FileRepository
}

func NewFileService(fileRepository *repository.FileRepository) FileService {
	return &fileService{
		FileRepository: fileRepository,
	}
}

func (s *fileService) InitUploadFile(req *model.UploadTask) (task *model.UploadTask, err error) {
	// 判断任务是否已经存在
	task, err = s.FileRepository.GetUploadTaskByHash(req.FileHash)
	if err != nil {
		// 任务不存在
		err = s.FileRepository.CreateUploadTask(req)
		return req, err
	}
	// 任务已经存在
	return task, nil
}

func (s *fileService) UploadFileChunk() error {
	return nil
}