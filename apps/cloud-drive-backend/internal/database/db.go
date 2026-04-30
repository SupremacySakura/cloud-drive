package database

import (
	"time"

	"cloud-drive-backend/internal/config"
	"cloud-drive-backend/internal/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	dsn := cfg.GetDSN()

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("数据库连接失败")
	}
	log.Info().Msg("数据库连接成功")
    // 配置连接池参数
    if sqlDB, err := DB.DB(); err == nil {
        sqlDB.SetMaxOpenConns(100)
        sqlDB.SetMaxIdleConns(10)
        sqlDB.SetConnMaxLifetime(time.Hour)
	} else {
		log.Warn().Err(err).Msg("无法获取底层数据库对象以设置连接池")
	}
	if err := Migrate(); err != nil {
		log.Fatal().Err(err).Msg("数据库迁移失败")
	}
	log.Info().Msg("数据库迁移成功")
}
