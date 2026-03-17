package db

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:123456123456@tcp(127.0.0.1:3306)/cloud-drive?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB,err = gorm.Open(mysql.Open(dsn),&gorm.Config{

	})
	if err != nil {
		log.Fatalf("数据库连接失败： %v",err)
	}
	log.Println("数据库连接成功")
	if err := Migrate(); err != nil {
		log.Fatalf("数据库迁移失败： %v",err)
	}
	log.Println("数据库迁移成功")
}