package utils

import (
	"fmt"
	pb "github.com/watchlist-kata/protos/subscription"
	"log"
	"net"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/watchlist-kata/subscription/api/server"
	"github.com/watchlist-kata/subscription/internal/config"
	"github.com/watchlist-kata/subscription/internal/service"
)

// SetupDatabase настраивает подключение к базе данных
func SetupDatabase(cfg *config.Config) (*gorm.DB, error) {
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

// StartGrpcServer запускает gRPC-сервер
func StartGrpcServer(cfg *config.Config, subscriptionService service.SubscriptionService) error {
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
