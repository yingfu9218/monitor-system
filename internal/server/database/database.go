package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/monitor-system/internal/server/model"
)

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS servers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		ip TEXT NOT NULL,
		os TEXT,
		location TEXT,
		status TEXT DEFAULT 'offline',
		last_heartbeat DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		cpu REAL,
		memory REAL,
		disk_read REAL,
		disk_write REAL,
		network_in REAL,
		network_out REAL,
		FOREIGN KEY (server_id) REFERENCES servers(id)
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_server_time ON metrics(server_id, timestamp DESC);

	CREATE TABLE IF NOT EXISTS server_info (
		server_id TEXT PRIMARY KEY,
		cpu_cores INTEGER,
		total_memory INTEGER,
		used_memory INTEGER,
		uptime INTEGER,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id)
	);

	CREATE TABLE IF NOT EXISTS disks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT NOT NULL,
		name TEXT,
		mount_point TEXT,
		fs_type TEXT,
		total_size INTEGER,
		used_size INTEGER,
		available_size INTEGER,
		usage_percent REAL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id)
	);

	CREATE TABLE IF NOT EXISTS processes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT NOT NULL,
		pid INTEGER,
		name TEXT,
		cpu REAL,
		memory REAL,
		username TEXT,
		status TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id)
	);

	CREATE INDEX IF NOT EXISTS idx_processes_server ON processes(server_id);

	CREATE TABLE IF NOT EXISTS network_interfaces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT NOT NULL,
		name TEXT,
		type TEXT,
		upload_speed REAL,
		download_speed REAL,
		total_upload INTEGER,
		total_download INTEGER,
		status TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id)
	);
	`

	_, err := db.Exec(schema)
	return err
}

func (db *DB) UpsertServer(server *model.Server) error {
	query := `
	INSERT INTO servers (id, name, ip, os, location, status, last_heartbeat, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		ip = excluded.ip,
		os = excluded.os,
		location = excluded.location,
		status = excluded.status,
		last_heartbeat = excluded.last_heartbeat,
		updated_at = excluded.updated_at
	`

	_, err := db.Exec(query, server.ID, server.Name, server.IP, server.OS,
		server.Location, server.Status, server.LastHeartbeat, time.Now())
	return err
}

func (db *DB) GetServers() ([]model.Server, error) {
	query := `SELECT id, name, ip, status, os, location, last_heartbeat, created_at, updated_at FROM servers ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []model.Server
	for rows.Next() {
		var s model.Server
		err := rows.Scan(&s.ID, &s.Name, &s.IP, &s.Status, &s.OS, &s.Location,
			&s.LastHeartbeat, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, nil
}

func (db *DB) GetServer(id string) (*model.Server, error) {
	query := `SELECT id, name, ip, status, os, location, last_heartbeat, created_at, updated_at
	          FROM servers WHERE id = ?`

	var s model.Server
	err := db.QueryRow(query, id).Scan(&s.ID, &s.Name, &s.IP, &s.Status, &s.OS,
		&s.Location, &s.LastHeartbeat, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}

	return &s, err
}

func (db *DB) InsertMetrics(metrics *model.Metrics) error {
	query := `
	INSERT INTO metrics (server_id, timestamp, cpu, memory, disk_read, disk_write, network_in, network_out)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, metrics.ServerID, metrics.Timestamp, metrics.CPU,
		metrics.Memory, metrics.DiskRead, metrics.DiskWrite, metrics.NetworkIn, metrics.NetworkOut)
	return err
}

func (db *DB) GetLatestMetrics(serverID string) (*model.Metrics, error) {
	query := `SELECT server_id, timestamp, cpu, memory, disk_read, disk_write, network_in, network_out
	          FROM metrics WHERE server_id = ? ORDER BY timestamp DESC LIMIT 1`

	var m model.Metrics
	err := db.QueryRow(query, serverID).Scan(&m.ServerID, &m.Timestamp, &m.CPU,
		&m.Memory, &m.DiskRead, &m.DiskWrite, &m.NetworkIn, &m.NetworkOut)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &m, err
}

func (db *DB) GetMetricsHistory(serverID string, duration time.Duration) ([]model.Metrics, error) {
	query := `SELECT server_id, timestamp, cpu, memory, disk_read, disk_write, network_in, network_out
	          FROM metrics WHERE server_id = ? AND timestamp >= ? ORDER BY timestamp ASC`

	since := time.Now().Add(-duration)
	rows, err := db.Query(query, serverID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []model.Metrics
	for rows.Next() {
		var m model.Metrics
		err := rows.Scan(&m.ServerID, &m.Timestamp, &m.CPU, &m.Memory,
			&m.DiskRead, &m.DiskWrite, &m.NetworkIn, &m.NetworkOut)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (db *DB) UpsertServerInfo(info *model.ServerInfo) error {
	query := `
	INSERT INTO server_info (server_id, cpu_cores, total_memory, used_memory, uptime, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(server_id) DO UPDATE SET
		cpu_cores = excluded.cpu_cores,
		total_memory = excluded.total_memory,
		used_memory = excluded.used_memory,
		uptime = excluded.uptime,
		updated_at = excluded.updated_at
	`

	_, err := db.Exec(query, info.ServerID, info.CPUCores, info.TotalMemory,
		info.UsedMemory, info.Uptime, time.Now())
	return err
}

func (db *DB) GetServerInfo(serverID string) (*model.ServerInfo, error) {
	query := `SELECT server_id, cpu_cores, total_memory, used_memory, uptime FROM server_info WHERE server_id = ?`

	var info model.ServerInfo
	err := db.QueryRow(query, serverID).Scan(&info.ServerID, &info.CPUCores,
		&info.TotalMemory, &info.UsedMemory, &info.Uptime)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &info, err
}

func (db *DB) ReplaceDisks(serverID string, disks []model.Disk) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old disks
	_, err = tx.Exec(`DELETE FROM disks WHERE server_id = ?`, serverID)
	if err != nil {
		return err
	}

	// Insert new disks
	stmt, err := tx.Prepare(`
		INSERT INTO disks (server_id, name, mount_point, fs_type, total_size, used_size, available_size, usage_percent, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, disk := range disks {
		_, err = stmt.Exec(serverID, disk.Name, disk.MountPoint, disk.FSType,
			disk.TotalSize, disk.UsedSize, disk.AvailableSize, disk.UsagePercent, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetDisks(serverID string) ([]model.Disk, error) {
	query := `SELECT name, mount_point, fs_type, total_size, used_size, available_size, usage_percent
	          FROM disks WHERE server_id = ?`

	rows, err := db.Query(query, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disks []model.Disk
	for rows.Next() {
		var d model.Disk
		err := rows.Scan(&d.Name, &d.MountPoint, &d.FSType, &d.TotalSize,
			&d.UsedSize, &d.AvailableSize, &d.UsagePercent)
		if err != nil {
			return nil, err
		}
		disks = append(disks, d)
	}

	return disks, nil
}

func (db *DB) ReplaceProcesses(serverID string, processes []model.Process) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old processes
	_, err = tx.Exec(`DELETE FROM processes WHERE server_id = ?`, serverID)
	if err != nil {
		return err
	}

	// Insert new processes
	stmt, err := tx.Prepare(`
		INSERT INTO processes (server_id, pid, name, cpu, memory, username, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, proc := range processes {
		_, err = stmt.Exec(serverID, proc.PID, proc.Name, proc.CPU,
			proc.Memory, proc.User, proc.Status, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetProcesses(serverID string, sortBy string, limit int) ([]model.Process, error) {
	orderBy := "cpu DESC"
	if sortBy == "memory" {
		orderBy = "memory DESC"
	}

	query := fmt.Sprintf(`SELECT pid, name, cpu, memory, username, status
	                      FROM processes WHERE server_id = ? ORDER BY %s LIMIT ?`, orderBy)

	rows, err := db.Query(query, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var processes []model.Process
	for rows.Next() {
		var p model.Process
		err := rows.Scan(&p.PID, &p.Name, &p.CPU, &p.Memory, &p.User, &p.Status)
		if err != nil {
			return nil, err
		}
		processes = append(processes, p)
	}

	return processes, nil
}

func (db *DB) ReplaceNetworkInterfaces(serverID string, interfaces []model.NetworkInterface) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old interfaces
	_, err = tx.Exec(`DELETE FROM network_interfaces WHERE server_id = ?`, serverID)
	if err != nil {
		return err
	}

	// Insert new interfaces
	stmt, err := tx.Prepare(`
		INSERT INTO network_interfaces (server_id, name, type, upload_speed, download_speed, total_upload, total_download, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, iface := range interfaces {
		_, err = stmt.Exec(serverID, iface.Name, iface.Type, iface.UploadSpeed,
			iface.DownloadSpeed, iface.TotalUpload, iface.TotalDownload, iface.Status, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetNetworkInterfaces(serverID string) ([]model.NetworkInterface, error) {
	query := `SELECT name, type, upload_speed, download_speed, total_upload, total_download, status
	          FROM network_interfaces WHERE server_id = ?`

	rows, err := db.Query(query, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interfaces []model.NetworkInterface
	for rows.Next() {
		var iface model.NetworkInterface
		err := rows.Scan(&iface.Name, &iface.Type, &iface.UploadSpeed, &iface.DownloadSpeed,
			&iface.TotalUpload, &iface.TotalDownload, &iface.Status)
		if err != nil {
			return nil, err
		}
		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

func (db *DB) UpdateServerStatus() error {
	// Set servers to warning if heartbeat > 30s, offline if > 60s
	now := time.Now()

	_, err := db.Exec(`
		UPDATE servers
		SET status = CASE
			WHEN julianday(?) - julianday(last_heartbeat) > (60.0 / 86400.0) THEN 'offline'
			WHEN julianday(?) - julianday(last_heartbeat) > (30.0 / 86400.0) THEN 'warning'
			ELSE 'online'
		END
		WHERE last_heartbeat IS NOT NULL
	`, now, now)

	return err
}

func (db *DB) CleanupOldData(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	_, err := db.Exec(`DELETE FROM metrics WHERE timestamp < ?`, cutoff)
	return err
}
