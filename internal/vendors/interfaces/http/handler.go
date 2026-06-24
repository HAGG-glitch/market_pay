package http



import (

	"net/http"

	"time"



	"github.com/gin-gonic/gin"

	"github.com/google/uuid"

	vendorapp "github.com/marketpay/backend/internal/vendors/application"

	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"

	shared "github.com/marketpay/backend/internal/shared/domain/model"

	apperrors "github.com/marketpay/backend/pkg/errors"

	"github.com/marketpay/backend/pkg/democtx"

	"github.com/marketpay/backend/pkg/middleware"

	"github.com/marketpay/backend/pkg/pagination"

)



// Swagger schema references.

var (

	_ vendormodel.Vendor

	_ vendormodel.MarketAssociation

)



// Handler handles vendor HTTP requests.

type Handler struct {

	vendorSvc *vendorapp.VendorService

}



// NewHandler constructs a vendor Handler.

func NewHandler(vendorSvc *vendorapp.VendorService) *Handler {

	return &Handler{vendorSvc: vendorSvc}

}



// RegisterRoutes mounts vendor routes.

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {

	vendors := rg.Group("/vendors")

	vendors.Use(auth, middleware.RequireActiveVendor())

	{

		vendors.POST("", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleFieldAgent), h.Create)

		vendors.GET("", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleLoanOfficer, shared.RoleFieldAgent), h.List)

		vendors.GET("/market-associations", h.ListMarketAssociations)

		vendors.GET("/:id", h.GetByID)

		vendors.GET("/:id/eligibility", h.CheckEligibility)

		vendors.PUT("/:id/kyc/approve", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleLoanOfficer), h.ApproveKYC)

		vendors.PUT("/:id/freeze", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin, shared.RoleSuperAdmin), h.Freeze)

		vendors.PUT("/:id/unfreeze", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin, shared.RoleSuperAdmin), h.Unfreeze)
		vendors.PUT("/:id/field-agent", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin, shared.RoleSuperAdmin), h.AssignFieldAgent)

	}

}



type createVendorRequest struct {

	FirstName           string    `json:"first_name" binding:"required"`

	LastName            string    `json:"last_name" binding:"required"`

	Phone               string    `json:"phone" binding:"required"`

	NationalIDNumber    string    `json:"national_id_number" binding:"required"`

	NationalIDType      string    `json:"national_id_type" binding:"required"`

	DateOfBirth         time.Time `json:"date_of_birth" binding:"required"`

	Address             string    `json:"address"`

	MarketAssociationID string    `json:"market_association_id" binding:"required"`

	BusinessName        string    `json:"business_name"`

	BusinessType        string    `json:"business_type"`

	GroupID             string    `json:"group_id"`

	FieldAgentID        string    `json:"field_agent_id"`

	PIN                 string    `json:"pin" binding:"required,len=4"`

}



type freezeVendorRequest struct {

	Reason string `json:"reason" binding:"required"`

}

type assignFieldAgentRequest struct {

	FieldAgentID string `json:"field_agent_id" binding:"required"`

}



// Create godoc

// @Summary Register a new vendor

// @Tags vendors

// @Security BearerAuth

// @Accept json

// @Produce json

// @Param body body createVendorRequest true "Vendor data"

// @Success 201 {object} vendormodel.Vendor

// @Router /vendors [post]

func (h *Handler) Create(c *gin.Context) {

	var req createVendorRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}



	maID, err := uuid.Parse(req.MarketAssociationID)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid market_association_id"})

		return

	}



	userIDStr := middleware.GetUserID(c)

	userID, _ := uuid.Parse(userIDStr)

	role := middleware.GetRole(c)

	isDemo := democtx.FromGin(c)



	var fieldAgentID *uuid.UUID

	if req.FieldAgentID != "" {

		if id, err := uuid.Parse(req.FieldAgentID); err == nil {

			fieldAgentID = &id

		}

	} else if role == shared.RoleFieldAgent {

		fieldAgentID = &userID

	}



	input := vendorapp.CreateVendorInput{

		UserID:              userID,

		FirstName:           req.FirstName,

		LastName:            req.LastName,

		Phone:               req.Phone,

		NationalIDNumber:    req.NationalIDNumber,

		NationalIDType:      req.NationalIDType,

		DateOfBirth:         req.DateOfBirth,

		Address:             req.Address,

		MarketAssociationID: maID,

		BusinessName:        req.BusinessName,

		BusinessType:        req.BusinessType,

		PIN:                 req.PIN,

		IsDemo:              isDemo,

		FieldAgentID:        fieldAgentID,

	}



	vendor, err := h.vendorSvc.Create(c.Request.Context(), input)

	if err != nil {

		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusCreated, vendor)

}



// List godoc

// @Summary List all vendors

// @Tags vendors

// @Security BearerAuth

// @Produce json

// @Success 200 {object} map[string]interface{}

// @Router /vendors [get]

func (h *Handler) List(c *gin.Context) {

	params := pagination.FromQuery(c)

	isDemo := democtx.FromGin(c)



	var fieldAgentID *uuid.UUID

	if middleware.GetRole(c) == shared.RoleFieldAgent {

		id, _ := uuid.Parse(middleware.GetUserID(c))

		fieldAgentID = &id

	}



	vendors, total, err := h.vendorSvc.List(c.Request.Context(), isDemo, fieldAgentID, params.Offset(), params.Limit)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, pagination.NewResponse(vendors, total, params))

}



// GetByID godoc

// @Summary Get vendor by ID

// @Tags vendors

// @Security BearerAuth

// @Param id path string true "Vendor ID"

// @Success 200 {object} vendormodel.Vendor

// @Router /vendors/{id} [get]

func (h *Handler) GetByID(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})

		return

	}

	vendor, err := h.vendorSvc.GetByID(c.Request.Context(), id)

	if err != nil {

		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, vendor)

}



// CheckEligibility godoc

// @Summary Check vendor loan eligibility

// @Tags vendors

// @Security BearerAuth

// @Param id path string true "Vendor ID"

// @Success 200 {object} map[string]interface{}

// @Router /vendors/{id}/eligibility [get]

func (h *Handler) CheckEligibility(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})

		return

	}

	if err := h.vendorSvc.CheckEligibility(c.Request.Context(), id); err != nil {

		c.JSON(apperrors.HTTPStatus(err), gin.H{"eligible": false, "reason": err.Error()})

		return

	}

	c.JSON(http.StatusOK, gin.H{"eligible": true})

}



// ApproveKYC godoc

// @Summary Approve vendor KYC

// @Tags vendors

// @Security BearerAuth

// @Param id path string true "Vendor ID"

// @Success 200 {object} vendormodel.Vendor

// @Router /vendors/{id}/kyc/approve [put]

func (h *Handler) ApproveKYC(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})

		return

	}

	approverID, _ := uuid.Parse(middleware.GetUserID(c))

	vendor, err := h.vendorSvc.ApproveKYC(c.Request.Context(), id, approverID)

	if err != nil {

		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, vendor)

}



func (h *Handler) Freeze(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})

		return

	}

	var req freezeVendorRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	actorID, _ := uuid.Parse(middleware.GetUserID(c))

	if err := h.vendorSvc.FreezeVendor(c.Request.Context(), id, actorID, middleware.GetRole(c), req.Reason); err != nil {

		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, gin.H{"message": "vendor frozen"})

}



func (h *Handler) Unfreeze(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})

		return

	}

	actorID, _ := uuid.Parse(middleware.GetUserID(c))

	if err := h.vendorSvc.UnfreezeVendor(c.Request.Context(), id, actorID, middleware.GetRole(c)); err != nil {

		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, gin.H{"message": "vendor unfrozen"})

}



func (h *Handler) AssignFieldAgent(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})

		return

	}

	var req assignFieldAgentRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	fieldAgentID, err := uuid.Parse(req.FieldAgentID)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid field_agent_id"})

		return

	}

	actorID, _ := uuid.Parse(middleware.GetUserID(c))

	if err := h.vendorSvc.AssignFieldAgent(c.Request.Context(), id, fieldAgentID, actorID, middleware.GetRole(c)); err != nil {

		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, gin.H{"message": "field agent assigned"})

}

// ListMarketAssociations godoc

// @Summary List all market associations

// @Tags vendors

// @Security BearerAuth

// @Success 200 {array} vendormodel.MarketAssociation

// @Router /vendors/market-associations [get]

func (h *Handler) ListMarketAssociations(c *gin.Context) {

	mas, err := h.vendorSvc.ListMarketAssociations(c.Request.Context())

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, mas)

}


