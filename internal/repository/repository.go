package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/watchlist-kata/protos/media"
	"github.com/watchlist-kata/protos/review"
	"github.com/watchlist-kata/protos/subscription"
	"github.com/watchlist-kata/protos/user"
	"github.com/watchlist-kata/protos/watchlist"
	"gorm.io/gorm"
)

// SubscriptionRepository представляет интерфейс репозитория для работы с подписками
type SubscriptionRepository interface {
	Subscribe(ctx context.Context, subscriberID uint, userID uint) error
	Unsubscribe(ctx context.Context, subscriberID uint, userID uint) error
	GetSubscriptions(ctx context.Context, userID uint) ([]uint, error)
	GetSubscribers(ctx context.Context, userID uint) ([]uint, error)
	IsSubscribed(ctx context.Context, subscriberID uint, userID uint) (bool, error)
	GetWatchlistsBySubscription(ctx context.Context, userID uint) ([]*subscription.WatchlistItem, error)
	GetReviewsBySubscription(ctx context.Context, userID uint) ([]*subscription.ReviewItem, error)
}

// PostgresSubscriptionRepository реализует SubscriptionRepository для PostgreSQL
type PostgresSubscriptionRepository struct {
	db              *gorm.DB
	logger          *slog.Logger
	mediaClient     media.MediaServiceClient
	reviewClient    review.ReviewServiceClient
	watchlistClient watchlist.WatchlistServiceClient
	userClient      user.UserServiceClient
}

// NewPostgresSubscriptionRepository создает новый экземпляр PostgresSubscriptionRepository
func NewPostgresSubscriptionRepository(db *gorm.DB, logger *slog.Logger, mediaAddr string, reviewAddr string, watchlistAddr string, userAddr string) (*PostgresSubscriptionRepository, error) {
	mediaConn, err := grpc.NewClient(
		mediaAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to media service", slog.Any("error", err))
		return nil, err
	}

	reviewConn, err := grpc.NewClient(
		reviewAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to review service", slog.Any("error", err))
		return nil, err
	}

	watchlistConn, err := grpc.NewClient(
		watchlistAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to watchlist service", slog.Any("error", err))
		return nil, err
	}

	userConn, err := grpc.NewClient(
		userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to user service", slog.Any("error", err))
		return nil, err
	}

	return &PostgresSubscriptionRepository{
		db:              db,
		logger:          logger,
		mediaClient:     media.NewMediaServiceClient(mediaConn),
		reviewClient:    review.NewReviewServiceClient(reviewConn),
		watchlistClient: watchlist.NewWatchlistServiceClient(watchlistConn),
		userClient:      user.NewUserServiceClient(userConn),
	}, nil
}

// Subscribe добавляет подписку на пользователя
func (r *PostgresSubscriptionRepository) Subscribe(ctx context.Context, subscriberID uint, userID uint) error {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, "Subscribe operation canceled", slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
	}

	subscription := &GormSubscription{
		SubscriberID: subscriberID,
		UserID:       userID,
		CreatedAt:    time.Now(),
	}

	if err := r.db.Create(subscription).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create subscription", slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, "subscription created successfully")
	return nil
}

// Unsubscribe удаляет подписку пользователя
func (r *PostgresSubscriptionRepository) Unsubscribe(ctx context.Context, subscriberID uint, userID uint) error {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, "Unsubscribe operation canceled", slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
	}

	if err := r.db.Where("subscriber_id = ? AND user_id = ?", subscriberID, userID).Delete(&GormSubscription{}).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to delete subscription", slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, "subscription deleted successfully")
	return nil
}

// GetSubscriptions получает список подписок пользователя
func (r *PostgresSubscriptionRepository) GetSubscriptions(ctx context.Context, userID uint) ([]uint, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, "GetSubscriptions operation canceled", slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	var subscriptions []GormSubscription
	if err := r.db.Where("subscriber_id = ?", userID).Find(&subscriptions).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to get subscriptions", slog.Any("error", err))
		return nil, err
	}

	subscribedToIDs := make([]uint, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		subscribedToIDs = append(subscribedToIDs, subscription.UserID)
	}

	r.logger.InfoContext(ctx, "subscriptions fetched successfully")
	return subscribedToIDs, nil
}

// GetSubscribers получает список подписчиков пользователя
func (r *PostgresSubscriptionRepository) GetSubscribers(ctx context.Context, userID uint) ([]uint, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, "GetSubscribers operation canceled", slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	var subscriptions []GormSubscription
	if err := r.db.Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to get subscribers", slog.Any("error", err))
		return nil, err
	}

	subscriberIDs := make([]uint, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		subscriberIDs = append(subscriberIDs, subscription.SubscriberID)
	}

	r.logger.InfoContext(ctx, "subscribers fetched successfully")
	return subscriberIDs, nil
}

// IsSubscribed проверяет, подписан ли пользователь на другого пользователя
func (r *PostgresSubscriptionRepository) IsSubscribed(ctx context.Context, subscriberID uint, userID uint) (bool, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, "IsSubscribed operation canceled", slog.Any("error", ctx.Err()))
		return false, ctx.Err()
	default:
	}

	var subscription GormSubscription
	if err := r.db.Where("subscriber_id = ? AND user_id = ?", subscriberID, userID).First(&subscription).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WarnContext(ctx, "subscription not found")
			return false, nil
		}
		r.logger.ErrorContext(ctx, "failed to check subscription", slog.Any("error", err))
		return false, err
	}

	r.logger.InfoContext(ctx, "subscription checked successfully")
	return true, nil
}

// WatchlistItem представляет элемент вотчлиста
type WatchlistItem struct {
	MediaID uint   `json:"media_id"`
	UserID  uint   `json:"user_id"`
	Title   string `json:"title"`
	Desc    string `json:"description"`
}

// ReviewItem представляет элемент отзыва
type ReviewItem struct {
	ReviewID uint   `json:"review_id"`
	UserID   uint   `json:"user_id"`
	Content  string `json:"content"`
	Rating   int    `json:"rating"`
}

// GetWatchlistsBySubscription получает вотчлисты пользователей, на которых подписан пользователь
func (r *PostgresSubscriptionRepository) GetWatchlistsBySubscription(ctx context.Context, userID uint) ([]*subscription.WatchlistItem, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, "GetWatchlistsBySubscription operation canceled", slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	subscribedToIDs, err := r.GetSubscriptions(ctx, userID)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get subscriptions", slog.Any("error", err))
		return nil, err
	}

	var watchlists []*subscription.WatchlistItem
	for _, subscribedToID := range subscribedToIDs {
		watchlistResponse, err := r.watchlistClient.GetWatchlist(ctx, &watchlist.GetWatchlistRequest{UserId: int64(subscribedToID)})
		if err != nil {
			r.logger.ErrorContext(ctx, "failed to get watchlist from watchlist service", slog.Any("error", err))
			return nil, err
		}

		for _, watchlistItem := range watchlistResponse.Watchlists {
			mediaResponse, err := r.mediaClient.GetMediaByID(ctx, &media.GetMediaByIDRequest{Id: watchlistItem.MediaId})
			if err != nil {
				r.logger.ErrorContext(ctx, "failed to get media info from media service", slog.Any("error", err))
				return nil, err
			}

			userResponse, err := r.userClient.GetByID(ctx, &user.GetUserRequest{Id: int64(subscribedToID)})
			if err != nil {
				r.logger.ErrorContext(ctx, "failed to get user info from user service", slog.Any("error", err))
				return nil, err
			}

			watchlistItemInfo := &subscription.WatchlistItem{
				MediaId:     watchlistItem.MediaId,
				UserId:      watchlistItem.UserId,
				UserName:    userResponse.User.Username,
				Title:       mediaResponse.NameEn,
				Description: mediaResponse.Description,
			}
			watchlists = append(watchlists, watchlistItemInfo)
		}
	}

	r.logger.InfoContext(ctx, "watchlists fetched successfully")
	return watchlists, nil
}

// GetReviewsBySubscription получает отзывы пользователей, на которых подписан пользователь
func (r *PostgresSubscriptionRepository) GetReviewsBySubscription(ctx context.Context, userID uint) ([]*subscription.ReviewItem, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, "GetReviewsBySubscription operation canceled", slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	subscribedToIDs, err := r.GetSubscriptions(ctx, userID)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get subscriptions", slog.Any("error", err))
		return nil, err
	}

	var reviews []*subscription.ReviewItem
	for _, subscribedToID := range subscribedToIDs {
		reviewResponse, err := r.reviewClient.GetByUser(ctx, &review.GetByUserRequest{UserId: int64(subscribedToID)})
		if err != nil {
			r.logger.ErrorContext(ctx, "failed to get reviews from review service", slog.Any("error", err))
			return nil, err
		}

		for _, reviewProto := range reviewResponse.Reviews {

			mediaResponse, err := r.mediaClient.GetMediaByID(ctx, &media.GetMediaByIDRequest{Id: reviewProto.MediaId})
			if err != nil {
				r.logger.ErrorContext(ctx, "failed to get media info from media service", slog.Any("error", err))
				return nil, err
			}

			userResponse, err := r.userClient.GetByID(ctx, &user.GetUserRequest{Id: int64(subscribedToID)})
			if err != nil {
				r.logger.ErrorContext(ctx, "failed to get user info from user service", slog.Any("error", err))
				return nil, err
			}

			reviewItem := &subscription.ReviewItem{
				ReviewId:  reviewProto.Id,
				UserId:    reviewProto.UserId,
				UserName:  userResponse.User.Username,
				Content:   reviewProto.Content,
				Rating:    reviewProto.Rating,
				MediaName: mediaResponse.NameEn,
				MediaYear: mediaResponse.Year,
			}
			reviews = append(reviews, reviewItem)
		}
	}

	r.logger.InfoContext(ctx, "reviews fetched successfully")
	return reviews, nil
}
