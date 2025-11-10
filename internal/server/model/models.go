package model

import "time"

type Server struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	IP            string    `json:"ip"`
	Status        string    `json:"status"`
	OS            string    `json:"os"`
	Location      string    `json:"location"`
	LastHeartbeat time.Time `json:"lastHeartbeat"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Metrics struct {
	ServerID   string    `json:"serverId"`
	Timestamp  time.Time `json:"timestamp"`
	CPU        float64   `json:"cpu"`
	Memory     float64   `json:"memory"`
	DiskRead   float64   `json:"diskRead"`
	DiskWrite  float64   `json:"diskWrite"`
	NetworkIn  float64   `json:"networkIn"`
	NetworkOut float64   `json:"networkOut"`
}

type ServerInfo struct {
	ServerID    string `json:"serverId"`
	CPUCores    int    `json:"cpuCores"`
	TotalMemory int64  `json:"totalMemory"`
	UsedMemory  int64  `json:"usedMemory"`
	Uptime      int64  `json:"uptime"`
}

type Disk struct {
	Name          string  `json:"name"`
	MountPoint    string  `json:"mountPoint"`
	FSType        string  `json:"fsType"`
	TotalSize     uint64  `json:"totalSize"`
	UsedSize      uint64  `json:"usedSize"`
	AvailableSize uint64  `json:"availableSize"`
	UsagePercent  float64 `json:"usagePercent"`
}

type Process struct {
	PID    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	User   string  `json:"user"`
	Status string  `json:"status"`
}

type NetworkInterface struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	UploadSpeed   float64 `json:"uploadSpeed"`
	DownloadSpeed float64 `json:"downloadSpeed"`
	TotalUpload   uint64  `json:"totalUpload"`
	TotalDownload uint64  `json:"totalDownload"`
	Status        string  `json:"status"`
}

type AgentReport struct {
	ServerID   string             `json:"serverId"`
	ServerName string             `json:"serverName,omitempty"` // Agent 配置中的服务器名称
	OS         string             `json:"os,omitempty"`         // 操作系统信息
	Location   string             `json:"location,omitempty"`   // 服务器位置
	Timestamp  time.Time          `json:"timestamp"`
	Metrics    Metrics            `json:"metrics"`
	Info       ServerInfo         `json:"info"`
	Disks      []Disk             `json:"disks"`
	Processes  []Process          `json:"processes"`
	Network    []NetworkInterface `json:"network"`
}
