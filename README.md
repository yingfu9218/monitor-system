# 服务器监控系统 (Server Monitoring System)

一个基于 Golang 的轻量级服务器监控系统，包含 API Server 和 Agent 两个组件。

app客户端应用请访问 https://github.com/yingfu9218/monitor 下载安装
## 系统架构

```
┌─────────────┐         ┌─────────────┐         ┌──────────────┐
│   前端UI    │ ◄─────► │  API Server │ ◄─────► │   Agent 1    │
│  (React)    │  REST   │  (Golang)   │  HTTP   │ (被监控服务器)│
└─────────────┘         └─────────────┘         └──────────────┘
                              │                   ┌──────────────┐
                              │◄─────────────────►│   Agent N    │
                              │      HTTP         │ (被监控服务器)│
                              ▼                   └──────────────┘
                        ┌──────────┐
                        │  数据库  │
                        │ (SQLite) │
                        └──────────┘
```

## 功能特性

### API Server
- RESTful API 接口
- SQLite 数据库存储
- API Key 认证
- 自动状态检测
- 历史数据查询
- 数据自动清理

### Agent
- 轻量级资源占用
- 系统指标采集
  - CPU 使用率
  - 内存使用情况
  - 磁盘 I/O
  - 网络流量
  - 进程信息
  - 磁盘分区
  - 网卡信息
- 定时上报（默认 5 秒）
- 跨平台支持（Linux, macOS, Windows）

## 快速开始

### 前置要求

- Go 1.21 或更高版本
- Git

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd monitor-system
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置文件

#### API Server 配置 (`configs/server-config.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "./data/monitor.db"

auth:
  api_key: "your-api-key-for-frontend"  # 修改为您的 API Key
  agent_key: "your-secret-agent-key"    # 修改为您的 Agent Key

data:
  retention_days: 30
  cleanup_interval: 24

logging:
  level: "info"
  file: "./logs/server.log"
```

#### Agent 配置 (`configs/agent-config.yaml`)

```yaml
server:
  id: "server-001"           # 服务器唯一ID
  name: "生产服务器 01"      # 服务器名称
  location: "北京"           # 服务器位置

api:
  endpoint: "http://localhost:8080"      # API Server 地址
  agent_key: "your-secret-agent-key"     # 与 Server 配置中的 agent_key 一致

reporting:
  interval: 5  # 上报间隔（秒）

logging:
  level: "info"
  file: "./logs/agent.log"
```

### 4. 编译

#### 编译 API Server

```bash
# Linux/macOS
go build -o bin/monitor-server ./cmd/server

# Windows
go build -o bin/monitor-server.exe ./cmd/server
```

#### 编译 Agent

```bash
# Linux/macOS
go build -o bin/monitor-agent ./cmd/agent

# Windows
go build -o bin/monitor-agent.exe ./cmd/agent
```

或使用提供的构建脚本：

```bash
chmod +x build.sh
./build.sh
```

### 5. 运行

#### 启动 API Server

```bash
./bin/monitor-server -config ./configs/server-config.yaml
```

服务器将在 `http://localhost:8080` 启动。

#### 启动 Agent（在被监控服务器上）

```bash
./bin/monitor-agent -config ./configs/agent-config.yaml
```

## API 文档

### 认证

所有前端 API 请求需要在 Header 中携带 API Key：

```
X-API-Key: your-api-key-for-frontend
```

Agent 请求需要携带 Agent Key：

```
X-Agent-Key: your-secret-agent-key
```

### 前端 API 接口

#### 1. 验证认证

```
POST /api/v1/auth/verify
Headers: X-API-Key: <api_key>

Response:
{
  "success": true,
  "message": "验证成功"
}
```

#### 2. 获取服务器列表

```
GET /api/v1/servers
Headers: X-API-Key: <api_key>

Response:
{
  "servers": [
    {
      "id": "server-001",
      "name": "生产服务器 01",
      "ip": "192.168.1.100",
      "status": "online",
      "os": "linux",
      "location": "北京",
      "lastHeartbeat": "2025-11-09T10:30:00Z",
      "currentMetrics": {
        "cpu": 45.2,
        "memory": 62.5,
        "network": 35.8
      }
    }
  ]
}
```

#### 3. 获取服务器详情

```
GET /api/v1/servers/:id
Headers: X-API-Key: <api_key>

Response:
{
  "server": {
    "id": "server-001",
    "name": "生产服务器 01",
    "ip": "192.168.1.100",
    "status": "online",
    "metrics": {
      "cpu": 45.2,
      "memory": 62.5,
      "diskRead": 25.3,
      "diskWrite": 18.7,
      "networkIn": 120.5,
      "networkOut": 85.3
    },
    "info": {
      "cpuCores": 8,
      "totalMemory": 16384,
      "usedMemory": 10240,
      "uptime": 1310400
    }
  }
}
```

#### 4. 获取历史数据

```
GET /api/v1/servers/:id/history?duration=20m
Headers: X-API-Key: <api_key>

Response:
{
  "history": [
    {
      "timestamp": "2025-11-09T10:10:00Z",
      "cpu": 42.1,
      "memory": 61.3,
      "diskRead": 23.5,
      "diskWrite": 16.2,
      "networkIn": 115.2,
      "networkOut": 82.1
    }
  ]
}
```

#### 5. 获取磁盘信息

```
GET /api/v1/servers/:id/disks
Headers: X-API-Key: <api_key>

Response:
{
  "disks": [
    {
      "name": "/dev/sda1",
      "mountPoint": "/",
      "fsType": "ext4",
      "totalSize": 512000,
      "usedSize": 296960,
      "availableSize": 215040,
      "usagePercent": 58.0
    }
  ]
}
```

#### 6. 获取进程列表

```
GET /api/v1/servers/:id/processes?sortBy=cpu&limit=20
Headers: X-API-Key: <api_key>

Response:
{
  "processes": [
    {
      "pid": 1234,
      "name": "nginx",
      "cpu": 45.2,
      "memory": 12.5,
      "user": "root",
      "status": "running"
    }
  ]
}
```

#### 7. 获取网卡信息

```
GET /api/v1/servers/:id/network
Headers: X-API-Key: <api_key>

Response:
{
  "interfaces": [
    {
      "name": "eth0",
      "type": "ethernet",
      "uploadSpeed": 85.5,
      "downloadSpeed": 120.3,
      "totalUpload": 1280000,
      "totalDownload": 3560000,
      "status": "up"
    }
  ]
}
```

## 部署指南

### 生产环境部署

#### API Server

1. 编译二进制文件
2. 修改配置文件中的认证密钥
3. 使用 systemd 或其他进程管理器运行

示例 systemd 服务文件 (`/etc/systemd/system/monitor-server.service`):

```ini
[Unit]
Description=Monitor API Server
After=network.target

[Service]
Type=simple
User=monitor
WorkingDirectory=/opt/monitor-system
ExecStart=/opt/monitor-system/bin/monitor-server -config /opt/monitor-system/configs/server-config.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable monitor-server
sudo systemctl start monitor-server
```

#### Agent

在每台需要监控的服务器上：

1. 复制编译好的 agent 二进制文件
2. 创建配置文件（修改 server.id, server.name, api.endpoint）
3. 使用 systemd 运行

示例 systemd 服务文件 (`/etc/systemd/system/monitor-agent.service`):

```ini
[Unit]
Description=Monitor Agent
After=network.target

[Service]
Type=simple
User=monitor
WorkingDirectory=/opt/monitor-agent
ExecStart=/opt/monitor-agent/bin/monitor-agent -config /opt/monitor-agent/configs/agent-config.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable monitor-agent
sudo systemctl start monitor-agent
```

## 前端集成

前端 React Native 应用位于项目根目录下。

### 配置

在应用的设置页面中配置：

- API 地址：API Server 的地址（如 `http://your-server-ip`）
- API 端口：API Server 的端口（默认 `8080`）
- API 密钥：在 `server-config.yaml` 中配置的 `api_key`

### 启动前端

```bash
cd ../monitor  # 回到 React Native 项目目录
npm install
npm start
```

## 故障排查

### API Server 无法启动

- 检查端口是否被占用
- 检查数据库文件路径是否有写权限
- 查看日志文件 `./logs/server.log`

### Agent 无法连接

- 检查 API Server 地址是否正确
- 检查 Agent Key 是否与 Server 配置一致
- 检查网络连接和防火墙设置
- 查看日志文件 `./logs/agent.log`

### 服务器状态显示离线

- 检查 Agent 是否正在运行
- 检查 Agent 和 Server 之间的网络连接
- 查看最后心跳时间（超过 60 秒显示离线）

## 开发

### 项目结构

```
monitor-system/
├── cmd/
│   ├── server/          # API Server 入口
│   └── agent/           # Agent 入口
├── internal/
│   ├── server/
│   │   ├── handler/     # HTTP 处理器
│   │   ├── middleware/  # 中间件
│   │   ├── model/       # 数据模型
│   │   ├── database/    # 数据库操作
│   │   └── config/      # 配置
│   └── agent/
│       ├── collector/   # 数据采集器
│       ├── reporter/    # 数据上报
│       └── config/      # 配置
├── configs/             # 配置文件
├── data/                # 数据库文件
├── logs/                # 日志文件
├── go.mod
└── README.md
```

### 运行测试

```bash
go test ./...
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题，请提交 Issue 或联系维护者。
