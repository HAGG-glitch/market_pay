package postgres

import (
	"context"

	ussdmodel "github.com/marketpay/backend/internal/ussd/domain/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SessionRepo implements SessionRepository.
type SessionRepo struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) CreateOrUpdate(ctx context.Context, session *ussdmodel.USSDSession) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"menu_state", "state_data", "pin_verified", "last_input", "expires_at", "updated_at"}),
		}).
		Create(session).Error
}

func (r *SessionRepo) FindBySessionID(ctx context.Context, sessionID string) (*ussdmodel.USSDSession, error) {
	var session ussdmodel.USSDSession
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND is_active = true", sessionID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) Deactivate(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).
		Model(&ussdmodel.USSDSession{}).
		Where("session_id = ?", sessionID).
		Update("is_active", false).Error
}
