package errors

import (
	"errors"
	"fmt"
)

// 业务错误类型定义
// 这些错误用于在service层和handler层之间传递特定的业务错误信息
var (
	// 认证相关错误
	ErrUserNotFound     = errors.New("用户不存在")
	ErrInvalidPassword  = errors.New("密码错误")
	ErrUnauthorized     = errors.New("未授权访问")
	ErrInvalidToken     = errors.New("无效的令牌")
	
	// 文件相关错误
	ErrFileNotFound     = errors.New("文件不存在")
	ErrFolderNotFound   = errors.New("文件夹不存在")
	ErrPermissionDenied = errors.New("权限不足")
	ErrInvalidFileName  = errors.New("无效的文件名")
	
	// 存储相关错误
	ErrStorageQuotaExceeded = errors.New("存储空间不足")
	ErrChunkSizeMismatch    = errors.New("分片大小不匹配")
	ErrInvalidMIMEType      = errors.New("不支持的文件类型")
	
	// 分享相关错误
	ErrPublicShareNotFound  = errors.New("分享链接不存在")
	ErrPickupCodeExpired    = errors.New("取件码已过期")
	ErrPickupTargetNotFound = errors.New("取件目标不存在")
	ErrPickupEmptyFolder    = errors.New("文件夹为空")
)

// Wrap 包装错误，添加上下文信息
// 使用fmt.Errorf的%w动词来保持原始错误的可检测性
func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Wrapf 使用格式化字符串包装错误
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	context := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", context, err)
}

// Is 判断错误链中是否包含目标错误
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 将错误转换为特定类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Unwrap 解包错误，获取原始错误
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
