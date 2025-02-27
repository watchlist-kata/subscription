package main

import (
	"fmt"
	"log"

	"github.com/watchlist-kata/subscription/internal/config"
	"github.com/watchlist-kata/subscription/internal/repository"
	"github.com/watchlist-kata/subscription/internal/service"
	"github.com/watchlist-kata/subscription/pkg/logger"
	"github.com/watchlist-kata/subscription/pkg/utils"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к базе данных и инициализация репозитория и сервиса
	db, err := utils.SetupDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Инициализация логгера
	logg, err := logger.NewLogger(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.ServiceName, cfg.LogBufferSize)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if multiHandler, ok := logg.Handler().(*logger.MultiHandler); ok {
			multiHandler.CloseAll()
		}
	}()

	mediaAddr := fmt.Sprintf("%s:%s", cfg.MediaServiceHost, cfg.MediaServicePort)
	reviewAddr := fmt.Sprintf("%s:%s", cfg.ReviewServiceHost, cfg.ReviewServicePort)
	watchlistAddr := fmt.Sprintf("%s:%s", cfg.WatchlistServiceHost, cfg.WatchlistServicePort)
	userAddr := fmt.Sprintf("%s:%s", cfg.UserServiceHost, cfg.UserServicePort)

	// Инициализация репозитория и сервиса
	repo, err := repository.NewPostgresSubscriptionRepository(db, logg, mediaAddr, reviewAddr, watchlistAddr, userAddr)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	subscriptionService := service.NewSubscriptionService(repo, logg)

	// Запуск gRPC-сервера
	if err := utils.StartGrpcServer(cfg, subscriptionService); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
