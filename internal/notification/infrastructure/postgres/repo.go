package postgres

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"gorm.io/gorm"
)

// NotificationRepo implements NotificationRepository.
type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Save(ctx context.Context, n *shared.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *NotificationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string) error {
	updates := map[string]interface{}{"status": status}
	if errMsg != "" {
		updates["error"] = errMsg
	}
	return r.db.WithContext(ctx).
		Model(&shared.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}
