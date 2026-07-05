package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mw "github.com/marketpay/backend/pkg/middleware"
	"github.com/marketpay/backend/pkg/realtime"
	"github.com/marketpay/backend/pkg/response"
	"gorm.io/gorm"
)

// Handler exposes in-app notifications and SSE stream.
type Handler struct {
	db  *gorm.DB
	hub *realtime.Hub
}

func NewHandler(db *gorm.DB, hub *realtime.Hub) *Handler {
	return &Handler{db: db, hub: hub}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	n := rg.Group("/notifications")
	n.Use(auth)
	{
		n.GET("", h.List)
		n.GET("/stream", h.Stream)
		n.PUT("/:id/read", h.MarkRead)
	}
}

func (h *Handler) List(c *gin.Context) {
	userID := mw.GetUserID(c)
	isDemo := mw.IsDemoMode(c)

	type row struct {
		ID        uuid.UUID `json:"id"`
		EventType string    `json:"event_type"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		IsRead    bool      `json:"is_read"`
		CreatedAt time.Time `json:"created_at"`
	}
	var rows []row
	h.db.Raw(`SELECT id, event_type, title, body, is_read, created_at
		FROM in_app_notifications WHERE recipient_id = ? AND is_demo = ?
		ORDER BY created_at DESC LIMIT 50`, userID, isDemo).Scan(&rows)

	response.Success(c, rows, "")
}

func (h *Handler) MarkRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "The notification ID provided is not valid.")
		return
	}
	userID := mw.GetUserID(c)
	h.db.Exec(`UPDATE in_app_notifications SET is_read = true WHERE id = ? AND recipient_id = ?`, id, userID)
	response.Success(c, nil, "Notification marked as read.")
}

func (h *Handler) Stream(c *gin.Context) {
	userID := mw.GetUserID(c)
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	ch := h.hub.Subscribe(userID)
	defer h.hub.Unsubscribe(userID, ch)

	flusher, _ := c.Writer.(http.Flusher)
	fmt.Fprintf(c.Writer, ": connected\n\n")
	if flusher != nil {
		flusher.Flush()
	}

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
			if flusher != nil {
				flusher.Flush()
			}
		case <-c.Request.Context().Done():
			return
		case <-time.After(30 * time.Second):
			fmt.Fprintf(c.Writer, ": keepalive\n\n")
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}
