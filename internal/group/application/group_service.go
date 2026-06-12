package application

import (
	"context"

	"github.com/google/uuid"
	groupmodel "github.com/marketpay/backend/internal/group/domain/model"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
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
	List(ctx context.Context, isDemo bool, fieldAgentID *uuid.UUID, offset, limit int) ([]*groupmodel.Group, int64, error)
	LogFreezeHistory(ctx context.Context, entityType string, entityID, actorID uuid.UUID, action, reason, actorRole string, isDemo bool) error
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
	Name         string
	Description  string
	LeaderID     uuid.UUID
	FieldAgentID *uuid.UUID
	IsDemo       bool
}

// Create creates a new lending group.
func (s *GroupService) Create(ctx context.Context, input CreateGroupInput) (*groupmodel.Group, error) {
	group := &groupmodel.Group{
		Name:         input.Name,
		Description:  input.Description,
		Status:       groupmodel.GroupStatusActive,
		LeaderID:     input.LeaderID,
		FieldAgentID: input.FieldAgentID,
		IsDemo:       input.IsDemo,
	}

	if err := s.groups.Create(ctx, group); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	member := &groupmodel.GroupMember{
		GroupID:  group.ID,
		VendorID: input.LeaderID,
		IsLeader: true,
	}
	if err := s.groups.AddMember(ctx, member); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	_ = s.events.Publish(ctx, "GroupCreated", group.ID.String(), map[string]interface{}{
		"group_id": group.ID.String(),
		"name":     group.Name,
		"is_demo":  group.IsDemo,
	})

	s.log.Info("group created", zap.String("group_id", group.ID.String()))
	return group, nil
}

// AddMember adds a vendor to a group.
func (s *GroupService) AddMember(ctx context.Context, groupID, vendorID uuid.UUID) error {
	group, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return apperrors.ErrNotFound("group")
	}

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
func (s *GroupService) FreezeGroup(ctx context.Context, groupID uuid.UUID, actorID uuid.UUID, actorRole shared.Role, reason string) error {
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
		"is_demo":  group.IsDemo,
	})
	_ = s.events.Publish(ctx, "AccountFrozen", group.ID.String(), map[string]interface{}{
		"group_id": group.ID.String(),
		"reason":   reason,
		"is_demo":  group.IsDemo,
	})

	_ = s.groups.LogFreezeHistory(ctx, "group", groupID, actorID, "FREEZE", reason, string(actorRole), group.IsDemo)
	return nil
}

// UnfreezeGroup reactivates a frozen group.
func (s *GroupService) UnfreezeGroup(ctx context.Context, groupID uuid.UUID, actorID uuid.UUID, actorRole shared.Role) error {
	group, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return apperrors.ErrNotFound("group")
	}

	group.Unfreeze()
	if err := s.groups.Update(ctx, group); err != nil {
		return apperrors.ErrInternalServer(err)
	}

	_ = s.events.Publish(ctx, "AccountUnfrozen", group.ID.String(), map[string]interface{}{
		"group_id": group.ID.String(),
		"is_demo":  group.IsDemo,
	})

	_ = s.groups.LogFreezeHistory(ctx, "group", groupID, actorID, "UNFREEZE", "", string(actorRole), group.IsDemo)
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

// List returns paginated groups scoped by demo mode.
func (s *GroupService) List(ctx context.Context, isDemo bool, fieldAgentID *uuid.UUID, offset, limit int) ([]*groupmodel.Group, int64, error) {
	return s.groups.List(ctx, isDemo, fieldAgentID, offset, limit)
}
