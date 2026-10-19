package repository

import (
	"zhixu-backend/internal/model"
	"errors"

	"gorm.io/gorm"
)

type OperationLogRepository struct {
	DB *gorm.DB
}

func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{
		DB: db,
	}
}

// CreateOperationLog 写入一条操作日志
func (r *OperationLogRepository) CreateOperationLog(log *model.OperationLogModel) error {
	return r.DB.Create(log).Error
}

// GetOperationLogsByUserID 分页查询用户的操作日志（按时间倒序）
func (r *OperationLogRepository) GetOperationLogsByUserID(userID uint, page, pageSize int) ([]model.OperationLogModel, error) {
	var logs []model.OperationLogModel
	err := r.DB.Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error
	return logs, err
}

// CountOperationLogsByUserID 统计用户操作日志总数
func (r *OperationLogRepository) CountOperationLogsByUserID(userID uint) (int64, error) {
	var count int64
	err := r.DB.Model(&model.OperationLogModel{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// PruneOperationLogs 仅保留用户最近 keep 条日志，返回实际清理条数。
// 先按 (created_at DESC, id DESC) 找到第 keep+1 新的记录作为 cutoff，再删除比它更旧的全部日志；
// 单独查 cutoff 规避 MySQL 不支持 DELETE 子查询带 LIMIT 的限制。
func (r *OperationLogRepository) PruneOperationLogs(userID uint, keep int) (int64, error) {
	var cutoff model.OperationLogModel
	err := r.DB.Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Offset(keep).
		Limit(1).
		Select("id").
		Take(&cutoff).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	result := r.DB.Where("user_id = ? AND id < ?", userID, cutoff.ID).Delete(&model.OperationLogModel{})
	return result.RowsAffected, result.Error
}
