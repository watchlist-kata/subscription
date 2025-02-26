package service

import (
	"errors"
	"github.com/watchlist-kata/subscription/internal/repository"
)

// SubscriptionService представляет сервис для работы с подписками
type SubscriptionService interface {
	Subscribe(subscriberID uint, subscribeToID uint) error
	Unsubscribe(subscriberID uint, subscribeToID uint) error
	GetSubscriptions(userID uint) ([]uint, error)
	GetSubscribers(userID uint) ([]uint, error)
	IsSubscribed(subscriberID uint, subscribeToID uint) (bool, error)
}

// subscriptionService реализует SubscriptionService
type subscriptionService struct {
	repo repository.SubscriptionRepository
}

// NewSubscriptionService создает новый экземпляр SubscriptionService
func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{
		repo: repo,
	}
}

// Subscribe добавляет подписку пользователя на другого пользователя
func (s *subscriptionService) Subscribe(subscriberID uint, subscribeToID uint) error {
	// Проверка, что пользователь не подписывается сам на себя
	if subscriberID == subscribeToID {
		return errors.New("cannot subscribe to yourself")
	}

	// Проверка, существует ли уже такая подписка
	isSubscribed, err := s.IsSubscribed(subscriberID, subscribeToID)
	if err != nil {
		return err
	}
	if isSubscribed {
		return errors.New("subscription already exists")
	}

	return s.repo.Subscribe(subscriberID, subscribeToID)
}

// Unsubscribe удаляет подписку пользователя
func (s *subscriptionService) Unsubscribe(subscriberID uint, subscribeToID uint) error {
	// Проверка, существует ли подписка
	isSubscribed, err := s.IsSubscribed(subscriberID, subscribeToID)
	if err != nil {
		return err
	}
	if !isSubscribed {
		return errors.New("subscription does not exist")
	}

	return s.repo.Unsubscribe(subscriberID, subscribeToID)
}

// GetSubscriptions получает список подписок пользователя
func (s *subscriptionService) GetSubscriptions(userID uint) ([]uint, error) {
	return s.repo.GetSubscriptions(userID)
}

// GetSubscribers получает список подписчиков пользователя
func (s *subscriptionService) GetSubscribers(userID uint) ([]uint, error) {
	return s.repo.GetSubscribers(userID)
}

// IsSubscribed проверяет, подписан ли пользователь на другого пользователя
func (s *subscriptionService) IsSubscribed(subscriberID uint, subscribeToID uint) (bool, error) {
	return s.repo.IsSubscribed(subscriberID, subscribeToID)
}
