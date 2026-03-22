package repository

import (
	"cloud-drive-backend/internal/dto"
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

func (r *FileRepository) GetFileListByFolderIDAndUserID(folderID uint, userID uint) ([]model.FileModel, error) {
	var files []model.FileModel
	err := r.DB.Where("parent_id = ? AND user_id = ?", folderID, userID).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *FileRepository) GetFolderListByFolderIDAndUserID(folderID uint, userID uint) ([]model.FolderModel, error) {
	var folders []model.FolderModel
	err := r.DB.Where("parent_id = ? AND user_id = ?", folderID, userID).Find(&folders).Error
	if err != nil {
		return nil, err
	}
	return folders, nil
}

func (r *FileRepository) GetListByFolderIDAndUserID(folderID uint, userID uint, page, pageSize int) ([]dto.FileListItem, error) {
	var list []dto.FileListItem
	err := r.DB.Raw(`
	SELECT 
		id, 
		name, 
		'folder' as type,
		'' as file_type,
		0 as size,
		updated_at
	FROM folder_models
	WHERE parent_id = ? AND user_id = ?

	UNION ALL

	SELECT 
		id, 
		name, 
		'file' as type,
		type as file_type,
		size,
		updated_at
	FROM file_models
	WHERE folder_id = ? AND user_id = ?

	ORDER BY type DESC, name ASC
	LIMIT ? OFFSET ?
`, folderID, userID, folderID, userID, pageSize, (page-1)*pageSize).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *FileRepository) GetListCountByFolderIDAndUserID(folderID uint, userID uint) (int64, error) {
	var count int64
	err := r.DB.Raw(`
	SELECT COUNT(*)
	FROM folder_models
	WHERE parent_id = ? AND user_id = ?

	UNION ALL

	SELECT COUNT(*)
	FROM file_models
	WHERE folder_id = ? AND user_id = ?
`, folderID, userID, folderID, userID).Scan(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
