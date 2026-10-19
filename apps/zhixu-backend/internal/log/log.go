// Package log 提供结构化日志封装，基于 zerolog
package log

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger 是全局日志实例
var Logger zerolog.Logger

func init() {
	initLogger("", "")
}

// Init 使用配置初始化日志
// level 可以是: trace, debug, info, warn, error, fatal, panic
// appEnv 可以是: development, dev, production
func Init(level, appEnv string) {
	initLogger(level, appEnv)
}

func initLogger(level, appEnv string) {
	// 从环境变量或参数读取日志级别
	levelStr := strings.ToLower(level)
	if levelStr == "" {
		levelStr = strings.ToLower(os.Getenv("LOG_LEVEL"))
	}

	lvl := zerolog.InfoLevel
	switch levelStr {
	case "debug":
		lvl = zerolog.DebugLevel
	case "info":
		lvl = zerolog.InfoLevel
	case "warn", "warning":
		lvl = zerolog.WarnLevel
	case "error":
		lvl = zerolog.ErrorLevel
	case "fatal":
		lvl = zerolog.FatalLevel
	case "panic":
		lvl = zerolog.PanicLevel
	case "trace":
		lvl = zerolog.TraceLevel
	}
	zerolog.SetGlobalLevel(lvl)

	// 设置时间格式
	zerolog.TimeFieldFormat = time.RFC3339

	// 确定应用环境
	env := strings.ToLower(appEnv)
	if env == "" {
		env = strings.ToLower(os.Getenv("APP_ENV"))
	}

	// 创建 JSON 格式 logger，输出到标准输出
	Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		CallerWithSkipFrameCount(2).
		Logger()

	// 如果是开发环境，使用 ConsoleWriter 便于阅读
	if env == "development" || env == "dev" {
		Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().
			Timestamp().
			CallerWithSkipFrameCount(2).
			Logger()
	}

	// 替换全局 log
	log.Logger = Logger
}

// Debug 输出调试日志
func Debug() *zerolog.Event {
	return Logger.Debug()
}

// Info 输出信息日志
func Info() *zerolog.Event {
	return Logger.Info()
}

// Warn 输出警告日志
func Warn() *zerolog.Event {
	return Logger.Warn()
}

// Error 输出错误日志
func Error() *zerolog.Event {
	return Logger.Error()
}

// Fatal 输出致命错误日志并退出程序
func Fatal() *zerolog.Event {
	return Logger.Fatal()
}

// Panic 输出恐慌日志并 panic
func Panic() *zerolog.Event {
	return Logger.Panic()
}

// WithStr 创建带字符串字段的日志事件
func WithStr(key, val string) *zerolog.Logger {
	l := Logger.With().Str(key, val).Logger()
	return &l
}

// WithInt 创建带整数字段的日志事件
func WithInt(key string, val int) *zerolog.Logger {
	l := Logger.With().Int(key, val).Logger()
	return &l
}

// WithErr 创建带错误字段的日志事件
func WithErr(err error) *zerolog.Logger {
	l := Logger.With().Err(err).Logger()
	return &l
}
