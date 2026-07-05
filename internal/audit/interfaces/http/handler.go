package http

import (
	"time"

	"github.com/gin-gonic/gin"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/democtx"
	"github.com/marketpay/backend/pkg/middleware"
	"github.com/marketpay/backend/pkg/response"
	"gorm.io/gorm"
)

// Handler exposes audit log queries.
type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	rg.GET("/audit-logs", auth, middleware.RequireRoles(shared.RoleSuperAdmin), h.List)
}

func (h *Handler) List(c *gin.Context) {
	isDemo := democtx.FromGin(c)
	actorID := c.Query("actor_id")
	resource := c.Query("resource")
	since := c.Query("since")

	type row struct {
		ID         string    `json:"id"`
		ActorID    string    `json:"actor_id"`
		ActorRole  string    `json:"actor_role"`
		Action     string    `json:"action"`
		Resource   string    `json:"resource"`
		ResourceID string    `json:"resource_id"`
		OldState   string    `json:"old_state,omitempty"`
		NewState   string    `json:"new_state,omitempty"`
		CreatedAt  time.Time `json:"created_at"`
	}

	query := h.db.Table("audit_logs").Where("deleted_at IS NULL")
	if actorID != "" {
		query = query.Where("actor_id = ?", actorID)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if since != "" {
		query = query.Where("created_at >= ?", since)
	}

	// Demo/live: filter via related entity when possible — audit logs join users for demo flag
	query = query.Where(`EXISTS (
		SELECT 1 FROM users u WHERE u.id = audit_logs.actor_id AND u.is_demo = ?
	) OR actor_id = '00000000-0000-0000-0000-000000000000'`, isDemo)

	var rows []row
	query.Order("created_at DESC").Limit(100).Find(&rows)
	response.Success(c, rows, "")
}
