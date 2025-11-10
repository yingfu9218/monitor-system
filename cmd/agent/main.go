package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/monitor-system/internal/agent/collector"
	"github.com/monitor-system/internal/agent/config"
	"github.com/monitor-system/internal/agent/reporter"
	"github.com/monitor-system/internal/server/model"
)

func main() {
	configPath := flag.String("config", "./configs/agent-config.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize collector and reporter
	col := collector.New()
	rep := reporter.New(cfg.API.Endpoint, cfg.API.AgentKey)

	log.Printf("Starting agent for server: %s (%s)", cfg.Server.Name, cfg.Server.ID)
	log.Printf("Reporting to: %s", cfg.API.Endpoint)

	// Get OS info
	osInfo := runtime.GOOS

	ticker := time.NewTicker(time.Duration(cfg.Reporting.Interval) * time.Second)
	defer ticker.Stop()

	// Send initial report immediately
	if err := sendReport(cfg, col, rep, osInfo); err != nil {
		log.Printf("Failed to send initial report: %v", err)
	}

	// Send periodic reports
	for range ticker.C {
		if err := sendReport(cfg, col, rep, osInfo); err != nil {
			log.Printf("Failed to send report: %v", err)
		} else {
			log.Printf("Report sent successfully")
		}
	}
}

func sendReport(cfg *config.Config, col *collector.Collector, rep *reporter.Reporter, osInfo string) error {
	// Collect all data
	metrics, err := col.CollectMetrics()
	if err != nil {
		return err
	}
	metrics.ServerID = cfg.Server.ID

	info, err := col.CollectServerInfo()
	if err != nil {
		return err
	}
	info.ServerID = cfg.Server.ID

	disks, err := col.CollectDisks()
	if err != nil {
		log.Printf("Failed to collect disks: %v", err)
		disks = []model.Disk{}
	}

	processes, err := col.CollectProcesses(20)
	if err != nil {
		log.Printf("Failed to collect processes: %v", err)
		processes = []model.Process{}
	}

	network, err := col.CollectNetwork()
	if err != nil {
		log.Printf("Failed to collect network: %v", err)
		network = []model.NetworkInterface{}
	}

	// Get hostname for server name if not set
	serverName := cfg.Server.Name
	if serverName == "" {
		hostname, err := os.Hostname()
		if err == nil {
			serverName = hostname
		} else {
			serverName = cfg.Server.ID
		}
	}

	// Get location from config or use default
	location := cfg.Server.Location
	if location == "" {
		location = "未知"
	}

	// Create report
	report := &model.AgentReport{
		ServerID:   cfg.Server.ID,
		ServerName: serverName, // 包含服务器名称
		OS:         osInfo,     // 包含操作系统信息
		Location:   location,   // 包含位置信息
		Timestamp:  time.Now(),
		Metrics:    *metrics,
		Info:       *info,
		Disks:      disks,
		Processes:  processes,
		Network:    network,
	}

	// Send report
	return rep.Report(report)
}
