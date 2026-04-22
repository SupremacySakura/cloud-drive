package repository

import (
	"cloud-drive-backend/internal/dto"
	"cloud-drive-backend/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *FileRepository) GetUploadTaskByHashAndUserIDAndFolderID(fileHash string, userID uint, folderID uint) (*model.UploadTask, error) {
	var task model.UploadTask
	err := r.DB.Where("file_hash = ? AND user_id = ? AND folder_id = ?", fileHash, userID, folderID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *FileRepository) GetFileByHashAndUserID(fileHash string, userID uint) (*model.FileModel, error) {
	var file model.FileModel
	err := r.DB.Where("file_hash = ? AND user_id = ?", fileHash, userID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) GetFileByFileIDAndUserID(fileID uint, userID uint) (*model.FileModel, error) {
	var file model.FileModel
	err := r.DB.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) CheckFileExistsInFolder(fileHash string, userID uint, folderID uint) (bool, error) {
	var count int64
	err := r.DB.Model(&model.FileModel{}).Where("file_hash = ? AND user_id = ? AND folder_id = ?", fileHash, userID, folderID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
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
	WHERE parent_id = ? AND user_id = ? AND deleted_at IS NULL

	UNION ALL

	SELECT 
		id, 
		name, 
		'file' as type,
		type as file_type,
		size,
		updated_at
	FROM file_models
	WHERE folder_id = ? AND user_id = ? AND deleted_at IS NULL

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
	SELECT 
		(SELECT COUNT(*) FROM folder_models WHERE parent_id = ? AND user_id = ? AND deleted_at IS NULL) +
		(SELECT COUNT(*) FROM file_models WHERE folder_id = ? AND user_id = ? AND deleted_at IS NULL)
`, folderID, userID, folderID, userID).Scan(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *FileRepository) GetStorageUsedByUserID(userID uint) (uint64, error) {
	var used uint64
	err := r.DB.Raw(`
	SELECT COALESCE(SUM(size), 0)
	FROM file_models
	WHERE user_id = ? AND deleted_at IS NULL
`, userID).Scan(&used).Error
	if err != nil {
		return 0, err
	}
	return used, nil
}

func (r *FileRepository) GetFileStatsByUserID(userID uint) ([]dto.DashboardFileStatItem, error) {
	var list []dto.DashboardFileStatItem
	err := r.DB.Raw(`
	SELECT
		type,
		COUNT(*) AS count,
		COALESCE(SUM(size), 0) AS size
	FROM file_models
	WHERE user_id = ? AND deleted_at IS NULL
	GROUP BY type
`, userID).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *FileRepository) GetRecentActivitiesByUserID(userID uint, limit int) ([]dto.DashboardRecentActivityItem, error) {
	var list []dto.DashboardRecentActivityItem
	err := r.DB.Raw(`
	SELECT
		f.id,
		f.name,
		COALESCE(fd.name, '根目录') AS folder_name,
		f.type AS file_type,
		f.size,
		f.updated_at
	FROM file_models f
	LEFT JOIN folder_models fd ON f.folder_id = fd.id AND fd.deleted_at IS NULL
	WHERE f.user_id = ? AND f.deleted_at IS NULL
	ORDER BY f.updated_at DESC
	LIMIT ?
`, userID, limit).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *FileRepository) MakeDirectory(folderID uint, name string, userID uint) (uint, error) {
	folder := &model.FolderModel{
		ParentID: folderID,
		Name:     name,
		UserID:   userID,
	}
	err := r.DB.Create(folder).Error
	if err != nil {
		return 0, err
	}
	return folder.ID, nil
}

func (r *FileRepository) GetFolderByFolderIDAndUserID(folderID uint, userID uint) (*model.FolderModel, error) {
	var folder model.FolderModel
	err := r.DB.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *FileRepository) RenameFileByIDAndUserID(fileID uint, userID uint, name string) error {
	result := r.DB.Model(&model.FileModel{}).
		Where("id = ? AND user_id = ?", fileID, userID).
		Updates(map[string]interface{}{
			"name":       name,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *FileRepository) RenameFolderByIDAndUserID(folderID uint, userID uint, name string) error {
	result := r.DB.Model(&model.FolderModel{}).
		Where("id = ? AND user_id = ?", folderID, userID).
		Updates(map[string]interface{}{
			"name":       name,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *FileRepository) MoveFileByIDAndUserID(fileID uint, userID uint, targetFolderID uint) error {
	result := r.DB.Model(&model.FileModel{}).
		Where("id = ? AND user_id = ?", fileID, userID).
		Updates(map[string]interface{}{
			"folder_id":  targetFolderID,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *FileRepository) MoveFolderByIDAndUserID(folderID uint, userID uint, targetFolderID uint) error {
	result := r.DB.Model(&model.FolderModel{}).
		Where("id = ? AND user_id = ?", folderID, userID).
		Updates(map[string]interface{}{
			"parent_id":  targetFolderID,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *FileRepository) DeleteFileByIDAndUserID(fileID uint, userID uint) error {
	result := r.DB.Where("id = ? AND user_id = ?", fileID, userID).Delete(&model.FileModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *FileRepository) GetDirectChildFoldersByParentAndUserID(parentID uint, userID uint) ([]model.FolderModel, error) {
	var folders []model.FolderModel
	if err := r.DB.Where("parent_id = ? AND user_id = ?", parentID, userID).Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

func (r *FileRepository) DeleteFilesByFolderIDsAndUserID(folderIDs []uint, userID uint) error {
	if len(folderIDs) == 0 {
		return nil
	}
	return r.DB.Where("folder_id IN ? AND user_id = ?", folderIDs, userID).Delete(&model.FileModel{}).Error
}

func (r *FileRepository) DeleteFoldersByIDsAndUserID(folderIDs []uint, userID uint) error {
	if len(folderIDs) == 0 {
		return nil
	}
	return r.DB.Where("id IN ? AND user_id = ?", folderIDs, userID).Delete(&model.FolderModel{}).Error
}

func (r *FileRepository) CreatePickUpCode(code *model.PickUpCodeModel) (uint, error) {
	err := r.DB.Create(code).Error
	if err != nil {
		return 0, err
	}
	return code.ID, nil
}

func (r *FileRepository) GetPickUpCodeListByUserIDAndPage(userID uint, page, pageSize int) ([]model.PickUpCodeModel, error) {
	var codeModels []model.PickUpCodeModel
	err := r.DB.Where("user_id = ? LIMIT ? OFFSET ?", userID, pageSize, (page-1)*pageSize).Find(&codeModels).Error
	if err != nil {
		return nil, err
	}
	return codeModels, nil
}

func (r *FileRepository) GetPickUpCodeListCountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.DB.Model(&model.PickUpCodeModel{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *FileRepository) GetPickUpCodeByCode(code string) (*model.PickUpCodeModel, error) {
	var pickupCode model.PickUpCodeModel
	if err := r.DB.Where("code = ?", code).First(&pickupCode).Error; err != nil {
		return nil, err
	}
	return &pickupCode, nil
}

func (r *FileRepository) GetFileByID(fileID uint) (*model.FileModel, error) {
	var file model.FileModel
	if err := r.DB.Where("id = ?", fileID).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) CreatePublicShareLink(link *model.PublicShareLinkModel) error {
	return r.DB.Create(link).Error
}

func (r *FileRepository) GetPublicShareLinkByFileIDAndUserID(fileID uint, userID uint) (*model.PublicShareLinkModel, error) {
	var link model.PublicShareLinkModel
	if err := r.DB.Where("file_id = ? AND user_id = ?", fileID, userID).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *FileRepository) GetPublicShareLinkByToken(token string) (*model.PublicShareLinkModel, error) {
	var link model.PublicShareLinkModel
	if err := r.DB.Where("token = ?", token).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *FileRepository) DeletePublicShareLinkByFileIDAndUserID(fileID uint, userID uint) error {
	result := r.DB.Where("file_id = ? AND user_id = ?", fileID, userID).Delete(&model.PublicShareLinkModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *FileRepository) GetFolderByID(folderID uint) (*model.FolderModel, error) {
	var folder model.FolderModel
	if err := r.DB.Where("id = ?", folderID).First(&folder).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *FileRepository) GetChildrenByFolderID(folderID uint) ([]model.FolderModel, []model.FileModel, error) {
	var folders []model.FolderModel
	if err := r.DB.Where("parent_id = ?", folderID).Find(&folders).Error; err != nil {
		return nil, nil, err
	}

	var files []model.FileModel
	if err := r.DB.Where("folder_id = ?", folderID).Find(&files).Error; err != nil {
		return nil, nil, err
	}

	return folders, files, nil
}

func (r *FileRepository) IncrementDownloadAndMaybeExpire(codeID uint, now time.Time) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var code model.PickUpCodeModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", codeID).
			First(&code).Error; err != nil {
			return err
		}

		if code.Status != model.PickUpCodeStatusActive {
			return errors.New("pickup code expired")
		}
		if now.After(code.ExpireTime) || code.Download >= code.MaxDownload {
			code.Status = model.PickUpCodeStatusExpire
			if err := tx.Save(&code).Error; err != nil {
				return err
			}
			return errors.New("pickup code expired")
		}

		code.Download++
		if code.Download >= code.MaxDownload {
			code.Status = model.PickUpCodeStatusExpire
		}
		return tx.Save(&code).Error
	})
}
