package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/marketpay/backend/internal/group/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeActiveGroup(memberCount int) *model.Group {
	g := &model.Group{
		Status:   model.GroupStatusActive,
		LeaderID: uuid.New(),
	}
	for i := 0; i < memberCount; i++ {
		g.Members = append(g.Members, model.GroupMember{VendorID: uuid.New()})
	}
	return g
}

func TestGroup_MemberCount(t *testing.T) {
	g := makeActiveGroup(3)
	assert.Equal(t, 3, g.MemberCount())
}

func TestGroup_IsFull(t *testing.T) {
	full := makeActiveGroup(model.GroupMaxSize)
	assert.True(t, full.IsFull())

	notFull := makeActiveGroup(model.GroupMaxSize - 1)
	assert.False(t, notFull.IsFull())
}

func TestGroup_MeetsMinimumSize(t *testing.T) {
	ok := makeActiveGroup(model.GroupMinSize)
	assert.True(t, ok.MeetsMinimumSize())

	tooSmall := makeActiveGroup(model.GroupMinSize - 1)
	assert.False(t, tooSmall.MeetsMinimumSize())
}

func TestGroup_CanCreateLoan_Active(t *testing.T) {
	g := makeActiveGroup(model.GroupMinSize)
	err := g.CanCreateLoan()
	require.NoError(t, err)
}

func TestGroup_CanCreateLoan_Frozen(t *testing.T) {
	g := makeActiveGroup(model.GroupMinSize)
	g.Status = model.GroupStatusFrozen
	err := g.CanCreateLoan()
	assert.Error(t, err)
}

func TestGroup_CanCreateLoan_TooSmall(t *testing.T) {
	g := makeActiveGroup(model.GroupMinSize - 1)
	err := g.CanCreateLoan()
	assert.Error(t, err)
}

func TestGroup_AddMember_Success(t *testing.T) {
	g := makeActiveGroup(3)
	member, err := g.AddMember(uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, 4, g.MemberCount())
}

func TestGroup_AddMember_Full(t *testing.T) {
	g := makeActiveGroup(model.GroupMaxSize)
	_, err := g.AddMember(uuid.New())
	assert.Error(t, err)
}

func TestGroup_AddMember_Frozen(t *testing.T) {
	g := makeActiveGroup(3)
	g.Freeze("member defaulted")
	_, err := g.AddMember(uuid.New())
	assert.Error(t, err)
}

func TestGroup_Freeze_Unfreeze(t *testing.T) {
	g := makeActiveGroup(5)
	g.Freeze("member defaulted")
	assert.True(t, g.IsFrozen())
	assert.Equal(t, "member defaulted", g.FreezeReason)

	g.Unfreeze()
	assert.False(t, g.IsFrozen())
	assert.Empty(t, g.FreezeReason)
}

func TestGroup_HasMember(t *testing.T) {
	vendorID := uuid.New()
	g := &model.Group{
		Status:  model.GroupStatusActive,
		Members: []model.GroupMember{{VendorID: vendorID}},
	}
	assert.True(t, g.HasMember(vendorID))
	assert.False(t, g.HasMember(uuid.New()))
}
