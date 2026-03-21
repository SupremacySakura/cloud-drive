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

func (r *FileRepository) GetFileByHash(fileHash string) (*model.FileModel, error) {
	var file model.FileModel
	err := r.DB.Where("file_hash = ?", fileHash).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
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

func (r *FileRepository) GetUploadTaskByHashAndUserID(fileHash string, userID uint) (*model.UploadTask, error) {
	var task model.UploadTask
	err := r.DB.Where("file_hash = ? AND user_id = ?", fileHash, userID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *FileRepository) GetUploadTaskByID(id uint) (*model.UploadTask, error) {
	var task model.UploadTask
	err := r.DB.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *FileRepository) GetUploadTaskByIDAndUserID(id uint, userID uint) (*model.UploadTask, error) {
	var task model.UploadTask
	err := r.DB.Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *FileRepository) UpdateUploadTask(task *model.UploadTask) error {
	return r.DB.Save(task).Error
}