package repository

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

// SubscriptionRepository представляет интерфейс репозитория для работы с подписками
type SubscriptionRepository interface {
	Subscribe(subscriberID uint, userID uint) error
	Unsubscribe(subscriberID uint, userID uint) error
	GetSubscriptions(userID uint) ([]uint, error)
	GetSubscribers(userID uint) ([]uint, error)
	IsSubscribed(subscriberID uint, userID uint) (bool, error)
}

// PostgresSubscriptionRepository реализует SubscriptionRepository для PostgreSQL
type PostgresSubscriptionRepository struct {
	db *gorm.DB
}

// NewPostgresSubscriptionRepository создает новый экземпляр PostgresSubscriptionRepository
func NewPostgresSubscriptionRepository(db *gorm.DB) *PostgresSubscriptionRepository {
	return &PostgresSubscriptionRepository{db: db}
}

// Subscribe добавляет подписку на пользователя
func (r *PostgresSubscriptionRepository) Subscribe(subscriberID uint, userID uint) error {
	subscription := &GormSubscription{
		SubscriberID: subscriberID,
		UserID:       userID,
		CreatedAt:    time.Now(),
	}

	return r.db.Create(subscription).Error
}

// Unsubscribe удаляет подписку пользователя
func (r *PostgresSubscriptionRepository) Unsubscribe(subscriberID uint, userID uint) error {
	return r.db.Where("subscriber_id = ? AND user_id = ?", subscriberID, userID).Delete(&GormSubscription{}).Error
}

// GetSubscriptions получает список подписок пользователя
func (r *PostgresSubscriptionRepository) GetSubscriptions(userID uint) ([]uint, error) {
	var subscriptions []GormSubscription
	if err := r.db.Where("subscriber_id = ?", userID).Find(&subscriptions).Error; err != nil {
		return nil, err
	}

	subscribedToIDs := make([]uint, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		subscribedToIDs = append(subscribedToIDs, subscription.UserID)
	}

	return subscribedToIDs, nil
}

// GetSubscribers получает список подписчиков пользователя
func (r *PostgresSubscriptionRepository) GetSubscribers(userID uint) ([]uint, error) {
	var subscriptions []GormSubscription
	if err := r.db.Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
		return nil, err
	}

	subscriberIDs := make([]uint, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		subscriberIDs = append(subscriberIDs, subscription.SubscriberID)
	}

	return subscriberIDs, nil
}

// IsSubscribed проверяет, подписан ли пользователь на другого пользователя
func (r *PostgresSubscriptionRepository) IsSubscribed(subscriberID uint, userID uint) (bool, error) {
	var subscription GormSubscription
	if err := r.db.Where("subscriber_id = ? AND user_id = ?", subscriberID, userID).First(&subscription).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
