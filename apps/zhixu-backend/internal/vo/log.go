package vo

import (
	"zhixu-backend/internal/model"
	"time"
)

// OperationLogListItem 操作日志列表项（目标名称/目录为写入时的快照）
type OperationLogListItem struct {
	ID         uint                         `json:"id"`
	Action     model.OperationLogAction     `json:"action"`
	TargetType model.OperationLogTargetType `json:"target_type"`
	TargetID   uint                         `json:"target_id"`
	TargetName string                       `json:"target_name"`
	FolderName string                       `json:"folder_name"`
	Detail     string                       `json:"detail"`
	CreatedAt  time.Time                    `json:"created_at"`
}

// OperationLogListResp 操作日志分页响应
type OperationLogListResp struct {
	List  []OperationLogListItem `json:"list"`
	Total int64                  `json:"total"`
}
