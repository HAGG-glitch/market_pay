package model

import (
	"github.com/google/uuid"
	apperrors "github.com/marketpay/backend/pkg/errors"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

const (
	GroupMinSize = 5
	GroupMaxSize = 10
)

// GroupStatus represents the group lifecycle.
type GroupStatus string

const (
	GroupStatusActive  GroupStatus = "ACTIVE"
	GroupStatusFrozen  GroupStatus = "FROZEN"
	GroupStatusDissolved GroupStatus = "DISSOLVED"
)

// Group is the aggregate root for group lending.
type Group struct {
	shared.BaseModel
	Name        string      `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Description string      `gorm:"type:text" json:"description"`
	Status      GroupStatus `gorm:"type:varchar(50);not null;default:'ACTIVE'" json:"status"`
	LeaderID    uuid.UUID   `gorm:"type:uuid;not null" json:"leader_id"`
	FreezeReason string     `gorm:"type:text" json:"freeze_reason,omitempty"`
	Members     []GroupMember `gorm:"foreignKey:GroupID" json:"members,omitempty"`
}

// GroupMember links a vendor to a group.
type GroupMember struct {
	shared.BaseModel
	GroupID  uuid.UUID `gorm:"type:uuid;not null;index" json:"group_id"`
	VendorID uuid.UUID `gorm:"type:uuid;not null;index" json:"vendor_id"`
	IsLeader bool      `gorm:"default:false" json:"is_leader"`
	JoinedAt shared.BaseModel `gorm:"-" json:"-"`
}

// MemberCount returns the number of active members.
func (g *Group) MemberCount() int {
	return len(g.Members)
}

// IsFull checks if the group has reached maximum size.
func (g *Group) IsFull() bool {
	return g.MemberCount() >= GroupMaxSize
}

// MeetsMinimumSize checks if the group is large enough.
func (g *Group) MeetsMinimumSize() bool {
	return g.MemberCount() >= GroupMinSize
}

// IsFrozen checks if the group is currently frozen.
func (g *Group) IsFrozen() bool {
	return g.Status == GroupStatusFrozen
}

// CanCreateLoan checks if group members can apply for new loans.
func (g *Group) CanCreateLoan() error {
	if g.IsFrozen() {
		return apperrors.ErrGroupFrozen
	}
	if !g.MeetsMinimumSize() {
		return apperrors.ErrGroupMinSize
	}
	return nil
}

// AddMember attempts to add a member to the group.
func (g *Group) AddMember(vendorID uuid.UUID) (*GroupMember, error) {
	if g.IsFull() {
		return nil, apperrors.ErrGroupFull
	}
	if g.IsFrozen() {
		return nil, apperrors.ErrGroupFrozen
	}

	member := &GroupMember{
		GroupID:  g.ID,
		VendorID: vendorID,
		IsLeader: false,
	}
	g.Members = append(g.Members, *member)
	return member, nil
}

// Freeze freezes the group, preventing new loans.
func (g *Group) Freeze(reason string) {
	g.Status = GroupStatusFrozen
	g.FreezeReason = reason
}

// Unfreeze reactivates a frozen group.
func (g *Group) Unfreeze() {
	g.Status = GroupStatusActive
	g.FreezeReason = ""
}

// HasMember checks if a vendor is a group member.
func (g *Group) HasMember(vendorID uuid.UUID) bool {
	for _, m := range g.Members {
		if m.VendorID == vendorID {
			return true
		}
	}
	return false
}
