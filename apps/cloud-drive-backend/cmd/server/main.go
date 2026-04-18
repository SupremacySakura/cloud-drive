package main

import (
	"cloud-drive-backend/internal/database"
	"cloud-drive-backend/internal/router"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件（如果存在）
	// 环境变量优先级高于 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用环境变量或默认值")
	}

	database.InitDB()
	r := router.SetUpRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	log.Printf("服务器启动，端口：%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
