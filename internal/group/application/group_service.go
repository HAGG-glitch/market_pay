package application

import (
	"context"

	"github.com/google/uuid"
	groupmodel "github.com/marketpay/backend/internal/group/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
)

// GroupRepository defines persistence for groups.
type GroupRepository interface {
	Create(ctx context.Context, group *groupmodel.Group) error
	FindByID(ctx context.Context, id uuid.UUID) (*groupmodel.Group, error)
	FindByVendorID(ctx context.Context, vendorID uuid.UUID) (*groupmodel.Group, error)
	Update(ctx context.Context, group *groupmodel.Group) error
	AddMember(ctx context.Context, member *groupmodel.GroupMember) error
	RemoveMember(ctx context.Context, groupID, vendorID uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*groupmodel.Group, int64, error)
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload interface{}) error
}

// GroupService handles group lending operations.
type GroupService struct {
	groups GroupRepository
	events EventPublisher
	log    *logger.Logger
}

// NewGroupService constructs a GroupService.
func NewGroupService(groups GroupRepository, events EventPublisher, log *logger.Logger) *GroupService {
	return &GroupService{groups: groups, events: events, log: log}
}

// CreateGroupInput holds group creation data.
type CreateGroupInput struct {
	Name        string
	Description string
	LeaderID    uuid.UUID
}

// Create creates a new lending group.
func (s *GroupService) Create(ctx context.Context, input CreateGroupInput) (*groupmodel.Group, error) {
	group := &groupmodel.Group{
		Name:        input.Name,
		Description: input.Description,
		Status:      groupmodel.GroupStatusActive,
		LeaderID:    input.LeaderID,
	}

	if err := s.groups.Create(ctx, group); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	// Add leader as first member
	member := &groupmodel.GroupMember{
		GroupID:  group.ID,
		VendorID: input.LeaderID,
		IsLeader: true,
	}
	if err := s.groups.AddMember(ctx, member); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	s.log.Info("group created", zap.String("group_id", group.ID.String()))
	return group, nil
}

// AddMember adds a vendor to a group.
func (s *GroupService) AddMember(ctx context.Context, groupID, vendorID uuid.UUID) error {
	group, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return apperrors.ErrNotFound("group")
	}

	// Reload members
	if group.IsFull() {
		return apperrors.ErrGroupFull
	}
	if group.IsFrozen() {
		return apperrors.ErrGroupFrozen
	}

	member := &groupmodel.GroupMember{
		GroupID:  groupID,
		VendorID: vendorID,
		IsLeader: false,
	}
	return s.groups.AddMember(ctx, member)
}

// FreezeGroup freezes a group due to a member default.
func (s *GroupService) FreezeGroup(ctx context.Context, groupID uuid.UUID, reason string) error {
	group, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return apperrors.ErrNotFound("group")
	}

	group.Freeze(reason)
	if err := s.groups.Update(ctx, group); err != nil {
		return apperrors.ErrInternalServer(err)
	}

	_ = s.events.Publish(ctx, "GroupFrozen", group.ID.String(), map[string]interface{}{
		"group_id": group.ID.String(),
		"reason":   reason,
	})

	return nil
}

// GetByID retrieves a group by ID.
func (s *GroupService) GetByID(ctx context.Context, id uuid.UUID) (*groupmodel.Group, error) {
	group, err := s.groups.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.ErrNotFound("group")
	}
	return group, nil
}

// List returns paginated groups.
func (s *GroupService) List(ctx context.Context, offset, limit int) ([]*groupmodel.Group, int64, error) {
	return s.groups.List(ctx, offset, limit)
}
