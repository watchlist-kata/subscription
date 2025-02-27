package server

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/watchlist-kata/protos/subscription"
	"github.com/watchlist-kata/subscription/internal/service"
)

// GrpcSubscriptionServer реализует gRPC-сервис подписок
type GrpcSubscriptionServer struct {
	pb.UnimplementedSubscriptionServiceServer
	subscriptionService service.SubscriptionService
}

// NewGrpcSubscriptionServer создает новый экземпляр gRPC-сервера подписок
func NewGrpcSubscriptionServer(subscriptionService service.SubscriptionService) *GrpcSubscriptionServer {
	return &GrpcSubscriptionServer{
		subscriptionService: subscriptionService,
	}
}

// Subscribe обрабатывает gRPC-запрос на подписку
func (s *GrpcSubscriptionServer) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	if req.SubscriberId == req.SubscribeToId {
		return nil, status.Errorf(codes.InvalidArgument, "cannot subscribe to yourself")
	}

	err := s.subscriptionService.Subscribe(ctx, uint(req.SubscriberId), uint(req.SubscribeToId))
	if err != nil {
		// Обработка ошибок
		if strings.Contains(err.Error(), "already exists") {
			return nil, status.Errorf(codes.AlreadyExists, "subscription already exists")
		}
		log.Printf("Failed to subscribe: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to process subscription")
	}

	return &pb.SubscribeResponse{Success: true}, nil
}

// Unsubscribe обрабатывает gRPC-запрос на отписку
func (s *GrpcSubscriptionServer) Unsubscribe(ctx context.Context, req *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
	err := s.subscriptionService.Unsubscribe(ctx, uint(req.SubscriberId), uint(req.UnsubscribeFromId))
	if err != nil {
		// Обработка ошибок
		if strings.Contains(err.Error(), "does not exist") {
			return nil, status.Errorf(codes.NotFound, "subscription does not exist")
		}
		log.Printf("Failed to unsubscribe: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to process unsubscription")
	}

	return &pb.UnsubscribeResponse{Success: true}, nil
}

// GetSubscriptions обрабатывает gRPC-запрос на получение списка подписок пользователя
func (s *GrpcSubscriptionServer) GetSubscriptions(ctx context.Context, req *pb.GetSubscriptionsRequest) (*pb.GetSubscriptionsResponse, error) {
	subscriptions, err := s.subscriptionService.GetSubscriptions(ctx, uint(req.UserId))
	if err != nil {
		log.Printf("Failed to get subscriptions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get subscriptions")
	}

	// Преобразование []uint в []int64 для ответа
	subscribedToIds := make([]int64, len(subscriptions))
	for i, id := range subscriptions {
		subscribedToIds[i] = int64(id)
	}

	return &pb.GetSubscriptionsResponse{SubscribedToIds: subscribedToIds}, nil
}

// GetSubscribers обрабатывает gRPC-запрос на получение списка подписчиков пользователя
func (s *GrpcSubscriptionServer) GetSubscribers(ctx context.Context, req *pb.GetSubscribersRequest) (*pb.GetSubscribersResponse, error) {
	subscribers, err := s.subscriptionService.GetSubscribers(ctx, uint(req.UserId))
	if err != nil {
		log.Printf("Failed to get subscribers: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get subscribers")
	}

	// Преобразование []uint в []int64 для ответа
	subscriberIds := make([]int64, len(subscribers))
	for i, id := range subscribers {
		subscriberIds[i] = int64(id)
	}

	return &pb.GetSubscribersResponse{SubscriberIds: subscriberIds}, nil
}

// CheckSubscription обрабатывает gRPC-запрос на проверку подписки
func (s *GrpcSubscriptionServer) CheckSubscription(ctx context.Context, req *pb.CheckSubscriptionRequest) (*pb.CheckSubscriptionResponse, error) {
	isSubscribed, err := s.subscriptionService.IsSubscribed(ctx, uint(req.SubscriberId), uint(req.SubscribeToId))
	if err != nil {
		log.Printf("Failed to check subscription: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to check subscription")
	}

	return &pb.CheckSubscriptionResponse{IsSubscribed: isSubscribed}, nil
}

// GetWatchlistsBySubscription обрабатывает gRPC-запрос на получение вотчлистов подписок
func (s *GrpcSubscriptionServer) GetWatchlistsBySubscription(ctx context.Context, req *pb.GetWatchlistsRequest) (*pb.GetWatchlistsResponse, error) {
	watchlists, err := s.subscriptionService.GetWatchlistsBySubscription(ctx, uint(req.UserId))
	if err != nil {
		log.Printf("Failed to get watchlists: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get watchlists")
	}

	return &pb.GetWatchlistsResponse{Watchlists: watchlists}, nil
}

// GetReviewsBySubscription обрабатывает gRPC-запрос на получение отзывов подписок
func (s *GrpcSubscriptionServer) GetReviewsBySubscription(ctx context.Context, req *pb.GetReviewsRequest) (*pb.GetReviewsResponse, error) {
	reviews, err := s.subscriptionService.GetReviewsBySubscription(ctx, uint(req.UserId))
	if err != nil {
		log.Printf("Failed to get reviews: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get reviews")
	}

	return &pb.GetReviewsResponse{Reviews: reviews}, nil
}
