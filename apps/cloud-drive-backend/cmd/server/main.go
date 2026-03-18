package main

import (
	"cloud-drive-backend/internal/database"
	"cloud-drive-backend/internal/router"
	"log"
	"os"
)

func main() {
	database.InitDB()
	r := router.SetUpRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
