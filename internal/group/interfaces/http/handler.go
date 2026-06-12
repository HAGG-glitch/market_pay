package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	groupapp "github.com/marketpay/backend/internal/group/application"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/democtx"
	"github.com/marketpay/backend/pkg/middleware"
	"github.com/marketpay/backend/pkg/pagination"
)

// Handler handles group HTTP requests.
type Handler struct {
	svc *groupapp.GroupService
}

func NewHandler(svc *groupapp.GroupService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/groups")
	g.Use(auth)
	{
		g.POST("", middleware.RequireRoles(shared.RoleVendor, shared.RoleAdmin, shared.RoleFieldAgent), h.Create)
		g.GET("", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleLoanOfficer, shared.RoleFieldAgent), h.List)
		g.GET("/:id", h.GetByID)
		g.POST("/:id/members", middleware.RequireRoles(shared.RoleVendor, shared.RoleAdmin, shared.RoleFieldAgent), h.AddMember)
		g.PUT("/:id/freeze", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin, shared.RoleSuperAdmin), h.Freeze)
		g.PUT("/:id/unfreeze", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin, shared.RoleSuperAdmin), h.Unfreeze)
	}
}

type createGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type addMemberRequest struct {
	VendorID string `json:"vendor_id" binding:"required"`
}

type freezeGroupRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderIDStr := middleware.GetUserID(c)
	leaderID, _ := uuid.Parse(leaderIDStr)
	isDemo := democtx.FromGin(c)

	var fieldAgentID *uuid.UUID
	if middleware.GetRole(c) == shared.RoleFieldAgent {
		fieldAgentID = &leaderID
	}

	group, err := h.svc.Create(c.Request.Context(), groupapp.CreateGroupInput{
		Name:         req.Name,
		Description:  req.Description,
		LeaderID:     leaderID,
		FieldAgentID: fieldAgentID,
		IsDemo:       isDemo,
	})
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, group)
}

func (h *Handler) List(c *gin.Context) {
	params := pagination.FromQuery(c)
	isDemo := democtx.FromGin(c)

	var fieldAgentID *uuid.UUID
	if middleware.GetRole(c) == shared.RoleFieldAgent {
		id, _ := uuid.Parse(middleware.GetUserID(c))
		fieldAgentID = &id
	}

	groups, total, err := h.svc.List(c.Request.Context(), isDemo, fieldAgentID, params.Offset(), params.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResponse(groups, total, params))
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}
	group, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *Handler) AddMember(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	var req addMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vendorID, err := uuid.Parse(req.VendorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor_id"})
		return
	}

	if err := h.svc.AddMember(c.Request.Context(), groupID, vendorID); err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

func (h *Handler) Freeze(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	var req freezeGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID, _ := uuid.Parse(middleware.GetUserID(c))
	if err := h.svc.FreezeGroup(c.Request.Context(), groupID, actorID, middleware.GetRole(c), req.Reason); err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group frozen"})
}

func (h *Handler) Unfreeze(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	actorID, _ := uuid.Parse(middleware.GetUserID(c))
	if err := h.svc.UnfreezeGroup(c.Request.Context(), groupID, actorID, middleware.GetRole(c)); err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group unfrozen"})
}
