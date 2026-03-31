package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

type Handler struct {
	haEngine  ha.HAEngine
	fdmAgent  fdm.Agent
	vclsAgent vcls.Agent
}

func NewHandler(haEngine ha.HAEngine, fdmAgent fdm.Agent, vclsAgent vcls.Agent) *Handler {
	return &Handler{
		haEngine:  haEngine,
		fdmAgent:  fdmAgent,
		vclsAgent: vclsAgent,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.GET("/status", h.GetStatus)
		api.GET("/degradation", h.GetDegradation)
		api.GET("/tasks", h.GetTasks)
		api.POST("/evaluate/:vmid", h.Evaluate)
	}
}

func (h *Handler) GetStatus(c *gin.Context) {
	cv := h.fdmAgent.ClusterView()
	c.JSON(http.StatusOK, gin.H{
		"leader_id": cv.LeaderID,
		"nodes":     cv.Nodes,
	})
}

func (h *Handler) GetDegradation(c *gin.Context) {
	level := h.fdmAgent.LocalDegradationLevel()
	c.JSON(http.StatusOK, gin.H{
		"level": level,
	})
}

func (h *Handler) GetTasks(c *gin.Context) {
	tasks := h.haEngine.ActiveTasks()
	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

func (h *Handler) Evaluate(c *gin.Context) {
	vmid := c.Param("vmid")
	decision, err := h.haEngine.Evaluate(c.Request.Context(), vmid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, decision)
}

//Personal.AI order the ending