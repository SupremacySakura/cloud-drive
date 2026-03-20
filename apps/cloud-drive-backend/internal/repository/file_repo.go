package repository

import (
	"cloud-drive-backend/internal/model"
	"gorm.io/gorm"
)

type FileRepository struct {
	DB *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{
		DB: db,
	}
}

func (r *FileRepository) Create(file *model.FileModel) error {
	return r.DB.Create(file).Error
}

func (r *FileRepository) CreateUploadTask(task *model.UploadTask) error {
	return r.DB.Create(task).Error
}

func (r *FileRepository) GetUploadTaskByHash(fileHash string) (*model.UploadTask, error) {
	var task model.UploadTask
	err := r.DB.Where("file_hash = ?", fileHash).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}