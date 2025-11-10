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

#### 方式一：使用 build.sh 脚本（推荐用于本地开发）

```bash
chmod +x build.sh
./build.sh
```

#### 方式二：手动编译

**编译 API Server**

```bash
# Linux/macOS
go build -o bin/monitor-server ./cmd/server

# Windows
go build -o bin/monitor-server.exe ./cmd/server
```

**编译 Agent**

```bash
# Linux/macOS
go build -o bin/monitor-agent ./cmd/agent

# Windows
go build -o bin/monitor-agent.exe ./cmd/agent
```

#### 方式三：使用 GoReleaser（推荐用于发布）

GoReleaser 支持跨平台编译、自动打包和发布。

**支持的平台架构**
- Linux: amd64, arm64, armv7 (适用于群辉 DS214+ 等设备)
- macOS: amd64
- Windows: amd64, arm64

**安装 GoReleaser**

```bash
# 使用 Go 安装
go install github.com/goreleaser/goreleaser/v2@latest

# 或使用 Makefile
make install-goreleaser
```

**本地构建（测试 GoReleaser 配置）**

```bash
# 使用 GoReleaser 本地构建（包含所有平台）
make build-local

# 如果没有 ARM 交叉编译工具链，跳过 ARMv7 构建
make build-skip-armv7

# 或直接使用 goreleaser
goreleaser build --snapshot --clean
```

构建产物将位于 `dist/` 目录。

**Windows 平台说明**
- Windows 版本的二进制文件会自动添加 `.exe` 后缀
- 压缩包格式为 `.zip` 而不是 `.tar.gz`
- 文件名格式：`monitor-system_v1.0.0_windows_amd64.zip`

**创建发布版本**

```bash
# 1. 创建并推送 git tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# 2. 使用 GoReleaser 发布（会自动推送到 GitHub Releases）
make release
# 或创建快照版本（不需要 git tag）
make release-snapshot
```

**可用的 Make 命令**

```bash
make help              # 显示所有可用命令
make build             # 使用 build.sh 构建
make build-local       # 使用 GoReleaser 本地构建（所有平台）
make build-skip-armv7  # 使用 GoReleaser 构建（跳过 ARMv7）
make release           # 创建正式发布（需要 git tag）
make release-snapshot  # 创建快照发布（不需要 git tag）
make test              # 运行测试
make clean             # 清理构建产物
```

#### 方式四：手动为群辉 NAS 编译 ARMv7（如 DS214+）

```bash
# 编译 Server
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/monitor-server-armv7 ./cmd/server

# 编译 Agent
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/monitor-agent-armv7 ./cmd/agent
```

使用 GoReleaser 构建后，ARMv7 版本的文件名格式为：
- `monitor-system_v1.0.0_linux_armv7.tar.gz`

### 5. 运行

#### 启动 API Server

**Linux/macOS:**
```bash
./bin/monitor-server -config ./configs/server-config.yaml
```

**Windows (PowerShell/CMD):**
```powershell
.\bin\monitor-server.exe -config .\configs\server-config.yaml
```

服务器将在 `http://localhost:8080` 启动。

#### 启动 Agent（在被监控服务器上）

**Linux/macOS:**
```bash
./bin/monitor-agent -config ./configs/agent-config.yaml
```

**Windows (PowerShell/CMD):**
```powershell
.\bin\monitor-agent.exe -config .\configs\agent-config.yaml
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

### 群辉 NAS 部署（如 DS214+）

群辉 NAS 使用 ARMv7 架构，需要使用对应的二进制文件。

#### 1. 获取 ARMv7 版本

使用 GoReleaser 发布后，下载 `monitor-system_v1.0.0_linux_armv7.tar.gz`，或手动编译：

```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o monitor-agent ./cmd/agent
```

#### 2. 上传到群辉

使用 SCP 或通过群辉文件管理器上传文件：

```bash
scp monitor-agent admin@your-nas-ip:/volume1/monitor/
```

#### 3. 创建配置文件

在群辉上创建配置文件 `/volume1/monitor/agent-config.yaml`：

```yaml
server:
  id: "synology-ds214"
  name: "群辉 DS214+"
  location: "家庭"

api:
  endpoint: "http://your-server-ip:8080"
  agent_key: "your-secret-agent-key"

reporting:
  interval: 5

logging:
  level: "info"
  file: "/volume1/monitor/logs/agent.log"
```

#### 4. 设置权限并运行

```bash
# SSH 登录到群辉
ssh admin@your-nas-ip

# 设置执行权限
chmod +x /volume1/monitor/monitor-agent

# 创建日志目录
mkdir -p /volume1/monitor/logs

# 运行
/volume1/monitor/monitor-agent -config /volume1/monitor/agent-config.yaml
```

#### 5. 配置自动启动（可选）

在群辉控制面板中：
1. 打开「任务计划」
2. 新建触发的任务 > 用户定义的脚本
3. 常规：设置任务名称，用户选择 root
4. 触发：开机
5. 任务设置：输入脚本

```bash
/volume1/monitor/monitor-agent -config /volume1/monitor/agent-config.yaml &
```

或者创建 systemd 服务（如果群辉支持）。

### Windows 部署

#### 使用 Windows 服务管理器

可以使用 [NSSM](https://nssm.cc/) 将 Agent 注册为 Windows 服务。

**1. 下载并安装 NSSM**

访问 https://nssm.cc/download 下载 NSSM。

**2. 安装服务**

```powershell
# 使用管理员权限打开 PowerShell

# 安装 Server 服务
nssm install MonitorServer "C:\monitor\monitor-server.exe" "-config" "C:\monitor\configs\server-config.yaml"

# 安装 Agent 服务
nssm install MonitorAgent "C:\monitor\monitor-agent.exe" "-config" "C:\monitor\configs\agent-config.yaml"
```

**3. 配置服务**

```powershell
# 设置服务描述
nssm set MonitorAgent Description "Monitor System Agent"

# 设置工作目录
nssm set MonitorAgent AppDirectory "C:\monitor"

# 设置服务启动类型为自动
nssm set MonitorAgent Start SERVICE_AUTO_START

# 启动服务
nssm start MonitorAgent
```

**4. 查看服务状态**

```powershell
# 查看服务状态
nssm status MonitorAgent

# 停止服务
nssm stop MonitorAgent

# 重启服务
nssm restart MonitorAgent

# 卸载服务
nssm remove MonitorAgent confirm
```

#### 使用任务计划程序（不推荐）

也可以使用 Windows 任务计划程序在开机时启动，但不如服务方式稳定。

**创建开机启动任务：**
1. 打开「任务计划程序」
2. 创建基本任务
3. 触发器选择「计算机启动时」
4. 操作选择「启动程序」
5. 程序路径：`C:\monitor\monitor-agent.exe`
6. 添加参数：`-config C:\monitor\configs\agent-config.yaml`

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
