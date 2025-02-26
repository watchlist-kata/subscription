package repository

import (
	"time"
)

// GormSubscription представляет модель подписки в базе данных
type GormSubscription struct {
	ID           uint `gorm:"primaryKey"`
	SubscriberID uint `gorm:"column:subscriber_id"`
	UserID       uint `gorm:"column:user_id"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName возвращает имя таблицы для модели GormSubscription
func (GormSubscription) TableName() string {
	return "subscription"
}
