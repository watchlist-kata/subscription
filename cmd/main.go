package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "github.com/watchlist-kata/protos/subscription"
	"github.com/watchlist-kata/subscription/api/server"
	"github.com/watchlist-kata/subscription/internal/config"
	"github.com/watchlist-kata/subscription/internal/repository"
	"github.com/watchlist-kata/subscription/internal/service"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()

	// Подключение к базе данных
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Инициализация репозитория и сервиса
	repo := repository.NewPostgresSubscriptionRepository(db)
	subscriptionService := service.NewSubscriptionService(repo)

	// Запуск gRPC-сервера
	if err := startGrpcServer(cfg, subscriptionService); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}

// setupDatabase настраивает подключение к базе данных
func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// startGrpcServer запускает gRPC-сервер
func startGrpcServer(cfg *config.Config, subscriptionService service.SubscriptionService) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	subscriptionServer := server.NewGrpcSubscriptionServer(subscriptionService)
	pb.RegisterSubscriptionServiceServer(grpcServer, subscriptionServer)

	log.Printf("Starting gRPC server on port %s...", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
