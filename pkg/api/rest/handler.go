package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/turtacn/hci-vcls/internal/app"
)

type Handler struct {
	svc *app.Service
	log *zap.Logger
}

func NewHandler(svc *app.Service, log *zap.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.Use(h.loggingMiddleware())

	api := router.Group("/api/v1")
	{
		api.GET("/version", h.GetVersion)
		api.GET("/status", h.GetStatus)
		api.GET("/degradation", h.GetDegradation)
		api.POST("/ha/evaluate", h.EvaluateHA)
		api.GET("/ha/tasks", h.ListTasks)
	}
}

func (h *Handler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log requests and setup trace ID context if needed
		c.Next()
	}
}

func (h *Handler) GetVersion(c *gin.Context) {
	// Dummy version response, ldflags could be used in real cmd execution
	c.JSON(http.StatusOK, gin.H{
		"version": "1.0.0",
		"commit":  "unknown",
		"date":    "unknown",
	})
}

func (h *Handler) GetStatus(c *gin.Context) {
	status := h.svc.Status()
	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetDegradation(c *gin.Context) {
	status := h.svc.Status()
	c.JSON(http.StatusOK, gin.H{
		"level": status.DegradationLevel,
	})
}

type EvaluateRequest struct {
	ClusterID string `json:"cluster_id" binding:"required"`
}

func (h *Handler) EvaluateHA(c *gin.Context) {
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := h.svc.EvaluateHA(c.Request.Context(), req.ClusterID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if plan == nil {
		c.JSON(http.StatusOK, gin.H{"message": "no plan created"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan_id": plan.ID,
		"message": "HA evaluation triggered",
	})
}

func (h *Handler) ListTasks(c *gin.Context) {
	// For testing purpose returning empty tasks. A real implementation would fetch from planRepo/taskRepo
	c.JSON(http.StatusOK, gin.H{
		"tasks": []interface{}{},
	})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	if err == app.ErrNotLeader {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "NOT_LEADER"})
		return
	}
	if err == app.ErrBelowThreshold {
		c.JSON(http.StatusOK, gin.H{"message": "Degradation is below threshold. No action required.", "code": "BELOW_THRESHOLD"})
		return
	}
	if err.Error() == "no healthy candidate host available" {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "code": "NO_CANDIDATE"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "INTERNAL_ERROR"})
}

//Personal.AI order the ending