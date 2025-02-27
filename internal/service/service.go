package service

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/watchlist-kata/protos/subscription"
	"github.com/watchlist-kata/subscription/internal/repository"
)

// SubscriptionService представляет сервис для работы с подписками
type SubscriptionService interface {
	Subscribe(ctx context.Context, subscriberID uint, subscribeToID uint) error
	Unsubscribe(ctx context.Context, subscriberID uint, subscribeToID uint) error
	GetSubscriptions(ctx context.Context, userID uint) ([]uint, error)
	GetSubscribers(ctx context.Context, userID uint) ([]uint, error)
	IsSubscribed(ctx context.Context, subscriberID uint, subscribeToID uint) (bool, error)
	GetWatchlistsBySubscription(ctx context.Context, userID uint) ([]*subscription.WatchlistItem, error)
	GetReviewsBySubscription(ctx context.Context, userID uint) ([]*subscription.ReviewItem, error)
}

// subscriptionService реализует SubscriptionService
type subscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *slog.Logger
}

// NewSubscriptionService создает новый экземпляр SubscriptionService
func NewSubscriptionService(repo repository.SubscriptionRepository, logger *slog.Logger) SubscriptionService {
	return &subscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *subscriptionService) checkContextCancelled(ctx context.Context, method string) error {
	select {
	case <-ctx.Done():
		s.logger.ErrorContext(ctx, fmt.Sprintf("%s operation canceled", method), slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
		return nil
	}
}

// Subscribe добавляет подписку пользователя на другого пользователя
func (s *subscriptionService) Subscribe(ctx context.Context, subscriberID uint, subscribeToID uint) error {
	if err := s.checkContextCancelled(ctx, "Subscribe"); err != nil {
		return status.Error(codes.Canceled, err.Error())
	}

	// Проверка, что пользователь не подписывается сам на себя
	if subscriberID == subscribeToID {
		s.logger.WarnContext(ctx, "cannot subscribe to yourself")
		return status.Errorf(codes.InvalidArgument, "Cannot subscribe to yourself")
	}

	// Проверка, существует ли уже такая подписка
	isSubscribed, err := s.IsSubscribed(ctx, subscriberID, subscribeToID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check subscription", slog.Any("error", err))
		return status.Errorf(codes.Internal, "Failed to check subscription: %v", err)
	}
	if isSubscribed {
		s.logger.WarnContext(ctx, "subscription already exists")
		return status.Errorf(codes.AlreadyExists, "Subscription already exists")
	}

	if err := s.repo.Subscribe(ctx, subscriberID, subscribeToID); err != nil {
		s.logger.ErrorContext(ctx, "failed to create subscription", slog.Any("error", err))
		return status.Errorf(codes.Internal, "Failed to create subscription: %v", err)
	}

	s.logger.InfoContext(ctx, "subscription created successfully")
	return nil
}

// Unsubscribe удаляет подписку пользователя
func (s *subscriptionService) Unsubscribe(ctx context.Context, subscriberID uint, subscribeToID uint) error {
	if err := s.checkContextCancelled(ctx, "Unsubscribe"); err != nil {
		return status.Error(codes.Canceled, err.Error())
	}

	// Проверка, существует ли подписка
	isSubscribed, err := s.IsSubscribed(ctx, subscriberID, subscribeToID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check subscription", slog.Any("error", err))
		return status.Errorf(codes.Internal, "Failed to check subscription: %v", err)
	}
	if !isSubscribed {
		s.logger.WarnContext(ctx, "subscription does not exist")
		return status.Errorf(codes.NotFound, "Subscription does not exist")
	}

	if err := s.repo.Unsubscribe(ctx, subscriberID, subscribeToID); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete subscription", slog.Any("error", err))
		return status.Errorf(codes.Internal, "Failed to delete subscription: %v", err)
	}

	s.logger.InfoContext(ctx, "subscription deleted successfully")
	return nil
}

// GetSubscriptions получает список подписок пользователя
func (s *subscriptionService) GetSubscriptions(ctx context.Context, userID uint) ([]uint, error) {
	if err := s.checkContextCancelled(ctx, "GetSubscriptions"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	subscribedToIDs, err := s.repo.GetSubscriptions(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get subscriptions", slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get subscriptions: %v", err)
	}

	s.logger.InfoContext(ctx, "subscriptions fetched successfully")
	return subscribedToIDs, nil
}

// GetSubscribers получает список подписчиков пользователя
func (s *subscriptionService) GetSubscribers(ctx context.Context, userID uint) ([]uint, error) {
	if err := s.checkContextCancelled(ctx, "GetSubscribers"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	subscriberIDs, err := s.repo.GetSubscribers(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get subscribers", slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get subscribers: %v", err)
	}

	s.logger.InfoContext(ctx, "subscribers fetched successfully")
	return subscriberIDs, nil
}

// IsSubscribed проверяет, подписан ли пользователь на другого пользователя
func (s *subscriptionService) IsSubscribed(ctx context.Context, subscriberID uint, subscribeToID uint) (bool, error) {
	if err := s.checkContextCancelled(ctx, "IsSubscribed"); err != nil {
		return false, status.Error(codes.Canceled, err.Error())
	}

	isSubscribed, err := s.repo.IsSubscribed(ctx, subscriberID, subscribeToID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check subscription", slog.Any("error", err))
		return false, status.Errorf(codes.Internal, "Failed to check subscription: %v", err)
	}

	s.logger.InfoContext(ctx, "subscription checked successfully")
	return isSubscribed, nil
}

// GetWatchlistsBySubscription получает вотчлисты пользователей, на которых подписан пользователь
func (s *subscriptionService) GetWatchlistsBySubscription(ctx context.Context, userID uint) ([]*subscription.WatchlistItem, error) {
	if err := s.checkContextCancelled(ctx, "GetWatchlistsBySubscription"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	watchlists, err := s.repo.GetWatchlistsBySubscription(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get watchlists", slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get watchlists: %v", err)
	}

	s.logger.InfoContext(ctx, "watchlists fetched successfully")
	return watchlists, nil
}

// GetReviewsBySubscription получает отзывы пользователей, на которых подписан пользователь
func (s *subscriptionService) GetReviewsBySubscription(ctx context.Context, userID uint) ([]*subscription.ReviewItem, error) {
	if err := s.checkContextCancelled(ctx, "GetReviewsBySubscription"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	reviews, err := s.repo.GetReviewsBySubscription(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get reviews", slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get reviews: %v", err)
	}

	s.logger.InfoContext(ctx, "reviews fetched successfully")
	return reviews, nil
}
