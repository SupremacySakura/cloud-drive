package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud-drive-backend/internal/config"
	"cloud-drive-backend/internal/database"
	"cloud-drive-backend/internal/log"
	"cloud-drive-backend/internal/repository"
	"cloud-drive-backend/internal/router"
	"cloud-drive-backend/internal/utils"
)

func main() {
	cfg := config.MustLoad()

	log.Init(cfg.LogLevel, cfg.AppEnv)
	database.InitDB(cfg)
	utils.InitJWT(cfg)
	r := router.SetUpRouter(cfg)

	port := cfg.Port
	if port == "" {
		port = "9000"
	}

	addr := ":" + port
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			<-ticker.C
			repo := repository.NewFileRepository(database.DB)
			if err := repo.DeleteExpiredChunks(24*time.Hour, cfg.ChunkStoragePath); err != nil {
				log.Error().Err(err).Msg("failed to delete expired upload chunks")
			}
		}
	}()

	go func() {
		log.Info().Str("port", port).Msg("服务器启动")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("服务器启动失败")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("正在关闭服务器")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("服务器关闭失败")
	}
	log.Info().Msg("服务器已退出")
}
