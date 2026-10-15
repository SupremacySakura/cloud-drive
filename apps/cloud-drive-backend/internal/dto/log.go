package dto

// GetOperationLogsReq 获取操作日志请求
type GetOperationLogsReq struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}
