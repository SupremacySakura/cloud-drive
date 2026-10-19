package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zhixu-backend/internal/config"
	"zhixu-backend/internal/database"
	"zhixu-backend/internal/handler"
	"zhixu-backend/internal/log"
	"zhixu-backend/internal/middleware"
	"zhixu-backend/internal/repository"
	"zhixu-backend/internal/router"
	"zhixu-backend/internal/service"
	"zhixu-backend/internal/utils"
)

func main() {
	cfg := config.MustLoad()

	log.Init(cfg.LogLevel, cfg.AppEnv)
	database.InitDB(cfg)
	utils.InitJWT(cfg)
	r := router.SetUpRouter(cfg)
	credentials, err := service.NewModelCredentials(cfg.ModelConfigEncryptionKey)
	if err != nil {
		log.Fatal().Msg("模型密钥加密配置无效")
	}
	router.RegisterAIConfigRouter(r, &service.AIConfigService{
		Store: &repository.AIConfigRepository{DB: database.DB}, Credentials: credentials,
	})
	chat := &handler.ChatHandler{
		Service: &service.ChatService{Store: &repository.AIConfigRepository{DB: database.DB}, Credentials: credentials, AllowedPrivateEndpoints: cfg.ModelAllowedPrivateEndpoints},
		AIURL:   cfg.ChatAIURL, InternalBackendURL: cfg.ChatBackendURL,
	}
	r.POST("/chat", middleware.AuthMiddleware(), chat.Send)
	r.POST("/internal/chat/v1/*endpoint", chat.Gateway)
	knowledgeService := service.NewKnowledgeService(repository.NewKnowledgeRepository(database.DB), cfg.FileStoragePath)
	router.RegisterKnowledgeRouter(r, knowledgeService)
	workerCtx, stopWorker := context.WithCancel(context.Background())
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		knowledgeService.Run(workerCtx)
	}()

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
	stopWorker()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("服务器关闭失败")
	}
	select {
	case <-workerDone:
	case <-ctx.Done():
		log.Warn().Msg("知识导入任务将在下次启动恢复")
	}
	log.Info().Msg("服务器已退出")
}
