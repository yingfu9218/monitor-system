package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/monitor-system/internal/server/database"
	"github.com/monitor-system/internal/server/model"
)

type Handler struct {
	db *database.DB
}

func New(db *database.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) VerifyAuth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "验证成功"})
}

func (h *Handler) GetServers(c *gin.Context) {
	servers, err := h.db.GetServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add current metrics for each server
	type ServerWithMetrics struct {
		model.Server
		CurrentMetrics *struct {
			CPU      float64 `json:"cpu"`
			Memory   float64 `json:"memory"`
			Network  float64 `json:"network"`
			Upload   float64 `json:"upload"`   // 上行速度 (MB/s)
			Download float64 `json:"download"` // 下行速度 (MB/s)
		} `json:"currentMetrics,omitempty"`
	}

	result := make([]ServerWithMetrics, 0, len(servers))
	for _, server := range servers {
		serverMetrics := ServerWithMetrics{Server: server}

		metrics, err := h.db.GetLatestMetrics(server.ID)
		if err == nil && metrics != nil {
			serverMetrics.CurrentMetrics = &struct {
				CPU      float64 `json:"cpu"`
				Memory   float64 `json:"memory"`
				Network  float64 `json:"network"`
				Upload   float64 `json:"upload"`
				Download float64 `json:"download"`
			}{
				CPU:      metrics.CPU,
				Memory:   metrics.Memory,
				Network:  (metrics.NetworkIn + metrics.NetworkOut) / 2,
				Upload:   metrics.NetworkOut, // 上行速度 = NetworkOut (MB/s)
				Download: metrics.NetworkIn,  // 下行速度 = NetworkIn (MB/s)
			}
		}

		result = append(result, serverMetrics)
	}

	c.JSON(http.StatusOK, gin.H{"servers": result})
}

func (h *Handler) GetServerDetail(c *gin.Context) {
	serverID := c.Param("id")

	server, err := h.db.GetServer(serverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	metrics, _ := h.db.GetLatestMetrics(serverID)
	info, _ := h.db.GetServerInfo(serverID)

	response := gin.H{
		"server": gin.H{
			"id":       server.ID,
			"name":     server.Name,
			"ip":       server.IP,
			"status":   server.Status,
			"os":       server.OS,
			"location": server.Location,
		},
	}

	if metrics != nil {
		response["server"].(gin.H)["metrics"] = gin.H{
			"cpu":        metrics.CPU,
			"memory":     metrics.Memory,
			"diskRead":   metrics.DiskRead,
			"diskWrite":  metrics.DiskWrite,
			"networkIn":  metrics.NetworkIn,
			"networkOut": metrics.NetworkOut,
		}
	}

	if info != nil {
		response["server"].(gin.H)["info"] = gin.H{
			"cpuCores":    info.CPUCores,
			"totalMemory": info.TotalMemory,
			"usedMemory":  info.UsedMemory,
			"uptime":      info.Uptime,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetHistory(c *gin.Context) {
	serverID := c.Param("id")
	durationStr := c.DefaultQuery("duration", "20m")

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration format"})
		return
	}

	history, err := h.db.GetMetricsHistory(serverID, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (h *Handler) GetDisks(c *gin.Context) {
	serverID := c.Param("id")

	disks, err := h.db.GetDisks(serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"disks": disks})
}

func (h *Handler) GetProcesses(c *gin.Context) {
	serverID := c.Param("id")
	sortBy := c.DefaultQuery("sortBy", "cpu")
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	processes, err := h.db.GetProcesses(serverID, sortBy, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"processes": processes})
}

func (h *Handler) GetNetwork(c *gin.Context) {
	serverID := c.Param("id")

	interfaces, err := h.db.GetNetworkInterfaces(serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"interfaces": interfaces})
}

func (h *Handler) DeleteServer(c *gin.Context) {
	serverID := c.Param("id")

	// Check if server exists
	_, err := h.db.GetServer(serverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Delete server and all related data
	if err := h.db.DeleteServer(serverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Server deleted successfully"})
}

func (h *Handler) AgentReport(c *gin.Context) {
	var report model.AgentReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update server status and heartbeat
	serverName := report.ServerName
	if serverName == "" {
		serverName = report.ServerID // 如果没有提供名称，使用 ID
	}

	serverOS := report.OS
	if serverOS == "" {
		serverOS = "Unknown" // 默认操作系统
	}

	serverLocation := report.Location
	if serverLocation == "" {
		serverLocation = "未知" // 默认位置
	}

	server := &model.Server{
		ID:            report.ServerID,
		Name:          serverName, // 使用 Agent 上报的服务器名称
		IP:            c.ClientIP(),
		OS:            serverOS,       // 使用 Agent 上报的操作系统
		Location:      serverLocation, // 使用 Agent 上报的位置
		Status:        "online",
		LastHeartbeat: report.Timestamp,
	}

	if err := h.db.UpsertServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert metrics
	if err := h.db.InsertMetrics(&report.Metrics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update server info
	if err := h.db.UpsertServerInfo(&report.Info); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update disks
	if len(report.Disks) > 0 {
		if err := h.db.ReplaceDisks(report.ServerID, report.Disks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Update processes
	if len(report.Processes) > 0 {
		if err := h.db.ReplaceProcesses(report.ServerID, report.Processes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Update network interfaces
	if len(report.Network) > 0 {
		if err := h.db.ReplaceNetworkInterfaces(report.ServerID, report.Network); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"nextReportInterval": 5,
	})
}
