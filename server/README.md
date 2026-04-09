# NexCtl Server

NexCtl Server 是控制面服务端，负责登录、节点纳管（控制台创建节点并下发固定凭据）、当前状态接收、Agent WebSocket 接入、在线状态管理，以及 task/file/upgrade/alert/audit 模块的统一接口预留。

## Run

```powershell
cd server
go run ./cmd/server -config configs/config.example.yaml
```

仓库根目录 `docker compose up server` 使用 `Dockerfile.dev`：**构建镜像时**已将 `./cmd/server` 编译为 `/usr/local/bin/nexctl-server`，容器内直接执行该二进制（不再使用 `go run`）。修改 Go 源码后需 **`docker compose build server`**（或 `up --build`）重新编译进镜像；`./server:/app` 挂载仍用于挂载配置文件等，可覆盖镜像内 `/app` 下的同名文件。

## Dependencies

- MySQL 8
- Redis 7

建议直接使用仓库根目录的 `docker compose up mysql redis`。Compose 中的 Redis 默认带口令（`REDIS_PASSWORD`，默认 `opspilot-dev`），在宿主机运行本服务时请设置环境变量 `OPSPILOT_REDIS_PASSWORD` 与之一致，或在配置文件中填写 `redis.password`。

## 环境变量（覆盖 YAML）

启动时若下列变量非空，会覆盖配置文件对应项（与 `internal/config/config.go` 中 `Load` 一致）：

| 变量 | 作用 |
|------|------|
| `OPSPILOT_SERVER_LISTEN_ADDR` | 监听地址，如 `:8080` |
| `OPSPILOT_SERVER_EXTERNAL_URL` | **控制面对外根 URL**（Agent 注册返回的 `ws_url` 由 `external_url` + `/api/v1/agents/ws` 拼接）。Docker 映射宿主机端口时请在 **仓库根目录** `.env` 中设置，或由 Compose 的 `environment` 传入。别名：`NEXCTL_SERVER_EXTERNAL_URL`。 |
| `OPSPILOT_MYSQL_DSN` | MySQL DSN |
| `OPSPILOT_REDIS_ADDR` / `OPSPILOT_REDIS_PASSWORD` / `OPSPILOT_REDIS_DB` | Redis |
| `OPSPILOT_JWT_SECRET` | JWT 密钥 |
| `OPSPILOT_WEBSOCKET_ALLOWED_ORIGINS` | 浏览器 WS Origin，逗号分隔 |

本地直接 `go run` 时，进程会依次尝试加载 **当前目录** 与 **上一级目录** 的 `.env`（`godotenv`），因此在 `server/` 下启动时仍能读取仓库根目录的 `OPSPILOT_SERVER_EXTERNAL_URL`；**不会**覆盖已在 shell 中设置的变量。

仓库根目录 `docker-compose.yml` 中 **`server` 服务**还使用（仅 Compose 插值，服务端进程不读取）：

- **`OPSPILOT_SERVER_HOST_PORT`**：宿主机映射端口，默认 `8080`，格式为 `<宿主机端口>:8080`。未设置 **`OPSPILOT_SERVER_EXTERNAL_URL`** 时，会默认 `http://127.0.0.1:<OPSPILOT_SERVER_HOST_PORT>`，与映射端口对齐。

## Seed Data

可直接使用根目录 `docker-compose.yml` 中的 seed：

- 用户名：`admin`
- 密码：`admin123`
- 在 Web 控制台「添加节点」可获取 `agent_id` / `agent_secret` / `node_key` 用于 Agent 配置

## API Paths

- `GET /healthz`
- `POST /api/v1/auth/login`
- `GET /api/v1/agents/ws`（鉴权：`X-NexCtl-Agent-Id`、`X-NexCtl-Agent-Secret` 请求头，勿使用 URL query）
- `GET /api/v1/me`
- `GET /api/v1/nodes`
- `GET /api/v1/nodes/{nodeID}`
- `GET /api/v1/nodes/{nodeID}/agent-credentials`
- `POST /api/v1/nodes/{nodeID}/runtime-state`
- `GET /api/v1/tasks`
- `GET /api/v1/files`
- `GET /api/v1/upgrades/releases`
- `GET /api/v1/alerts/rules`
- `GET /api/v1/alerts/events`
- `GET /api/v1/audit/logs`

模块类 `GET` 需 `modules:read`；节点列表与详情需 `nodes:read`；`POST .../runtime-state` 需 `nodes:write`。`admin` 等角色拥有全部权限，`viewer`/`readonly` 为只读组合。

## WebSocket Protocol

请求和响应统一使用同一包结构：

```json
{
  "type": "heartbeat",
  "request_id": "req-123",
  "timestamp": "2026-04-08T10:30:00Z",
  "payload": {}
}
```

已实现消息：

- `heartbeat`
- `runtime_state`
- `ack`
- `error`

预留消息：

- `task_dispatch`
- `file_dispatch`
- `upgrade_command`
