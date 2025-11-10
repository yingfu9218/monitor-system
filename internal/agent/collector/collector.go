package collector

import (
	"runtime"
	"time"

	"github.com/monitor-system/internal/server/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type Collector struct {
	lastNetIO   map[string]net.IOCountersStat
	lastDiskIO  map[string]disk.IOCountersStat
	lastTime    time.Time
	lastNetTime map[string]time.Time // 每个网卡独立的时间戳
}

func New() *Collector {
	return &Collector{
		lastNetIO:   make(map[string]net.IOCountersStat),
		lastDiskIO:  make(map[string]disk.IOCountersStat),
		lastTime:    time.Now(),
		lastNetTime: make(map[string]time.Time),
	}
}

func (c *Collector) CollectMetrics() (*model.Metrics, error) {
	metrics := &model.Metrics{
		Timestamp: time.Now(),
	}

	// CPU
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPU = cpuPercent[0]
	}

	// Memory
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		metrics.Memory = vmStat.UsedPercent
	}

	// Disk IO
	ioCounters, err := disk.IOCounters()
	if err == nil {
		now := time.Now()
		elapsed := now.Sub(c.lastTime).Seconds()

		var totalReadBytes, totalWriteBytes uint64
		for name, io := range ioCounters {
			if last, ok := c.lastDiskIO[name]; ok {
				totalReadBytes += io.ReadBytes - last.ReadBytes
				totalWriteBytes += io.WriteBytes - last.WriteBytes
			}
			c.lastDiskIO[name] = io
		}

		if elapsed > 0 {
			// MB/s
			metrics.DiskRead = float64(totalReadBytes) / elapsed / 1024 / 1024
			metrics.DiskWrite = float64(totalWriteBytes) / elapsed / 1024 / 1024
		}
	}

	// Network
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		now := time.Now()
		elapsed := now.Sub(c.lastTime).Seconds()

		if last, ok := c.lastNetIO["total"]; ok {
			if elapsed > 0 {
				// MB/s
				metrics.NetworkIn = float64(netIO[0].BytesRecv-last.BytesRecv) / elapsed / 1024 / 1024
				metrics.NetworkOut = float64(netIO[0].BytesSent-last.BytesSent) / elapsed / 1024 / 1024
			}
		}
		c.lastNetIO["total"] = netIO[0]
	}

	c.lastTime = time.Now()

	return metrics, nil
}

func (c *Collector) CollectServerInfo() (*model.ServerInfo, error) {
	info := &model.ServerInfo{}

	// CPU cores
	info.CPUCores = runtime.NumCPU()

	// Memory
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		info.TotalMemory = int64(vmStat.Total / 1024 / 1024) // MB
		info.UsedMemory = int64(vmStat.Used / 1024 / 1024)   // MB
	}

	// Uptime
	uptime, err := host.Uptime()
	if err == nil {
		info.Uptime = int64(uptime)
	}

	return info, nil
}

func (c *Collector) CollectDisks() ([]model.Disk, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var disks []model.Disk
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		disks = append(disks, model.Disk{
			Name:          partition.Device,
			MountPoint:    partition.Mountpoint,
			FSType:        partition.Fstype,
			TotalSize:     usage.Total / 1024 / 1024, // MB
			UsedSize:      usage.Used / 1024 / 1024,  // MB
			AvailableSize: usage.Free / 1024 / 1024,  // MB
			UsagePercent:  usage.UsedPercent,
		})
	}

	return disks, nil
}

func (c *Collector) CollectProcesses(limit int) ([]model.Process, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	type ProcessWithMetrics struct {
		proc   *process.Process
		cpu    float64
		memory float64
	}

	var processMetrics []ProcessWithMetrics
	for _, p := range procs {
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		memPercent, err := p.MemoryPercent()
		if err != nil {
			continue
		}

		processMetrics = append(processMetrics, ProcessWithMetrics{
			proc:   p,
			cpu:    cpuPercent,
			memory: float64(memPercent),
		})
	}

	// Sort by CPU usage (simple bubble sort for top processes)
	for i := 0; i < len(processMetrics)-1; i++ {
		for j := 0; j < len(processMetrics)-i-1; j++ {
			if processMetrics[j].cpu < processMetrics[j+1].cpu {
				processMetrics[j], processMetrics[j+1] = processMetrics[j+1], processMetrics[j]
			}
		}
	}

	// Get top N processes
	if len(processMetrics) > limit {
		processMetrics = processMetrics[:limit]
	}

	var processes []model.Process
	for _, pm := range processMetrics {
		name, _ := pm.proc.Name()
		username, _ := pm.proc.Username()
		status, _ := pm.proc.Status()

		processes = append(processes, model.Process{
			PID:    pm.proc.Pid,
			Name:   name,
			CPU:    pm.cpu,
			Memory: pm.memory,
			User:   username,
			Status: status[0],
		})
	}

	return processes, nil
}

func (c *Collector) CollectNetwork() ([]model.NetworkInterface, error) {
	netIO, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var interfaces []model.NetworkInterface

	for _, io := range netIO {
		// Skip loopback
		if io.Name == "lo" || io.Name == "lo0" {
			continue
		}

		var uploadSpeed, downloadSpeed float64

		// 使用每个网卡独立的时间戳
		if last, ok := c.lastNetIO[io.Name]; ok {
			if lastTime, hasTime := c.lastNetTime[io.Name]; hasTime {
				elapsed := now.Sub(lastTime).Seconds()
				if elapsed > 0 {
					uploadSpeed = float64(io.BytesSent-last.BytesSent) / elapsed / 1024 / 1024   // MB/s
					downloadSpeed = float64(io.BytesRecv-last.BytesRecv) / elapsed / 1024 / 1024 // MB/s
				}
			}
		}

		interfaces = append(interfaces, model.NetworkInterface{
			Name:          io.Name,
			Type:          "ethernet", // Could be improved with more detection
			UploadSpeed:   uploadSpeed,
			DownloadSpeed: downloadSpeed,
			TotalUpload:   io.BytesSent / 1024 / 1024, // MB
			TotalDownload: io.BytesRecv / 1024 / 1024, // MB
			Status:        "up",
		})

		// 更新该网卡的历史数据和时间戳
		c.lastNetIO[io.Name] = io
		c.lastNetTime[io.Name] = now
	}

	return interfaces, nil
}
