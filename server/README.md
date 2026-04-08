# NexCtl Server

NexCtl Server 是控制面服务端，负责登录、节点注册、当前状态接收、Agent WebSocket 接入、在线状态管理，以及 task/file/upgrade/alert/audit 模块的统一接口预留。

## Run

```powershell
cd server
go run ./cmd/server -config configs/config.example.yaml
```

## Dependencies

- MySQL 8
- Redis 7

建议直接使用仓库根目录的 `docker compose up mysql redis`。Compose 中的 Redis 默认带口令（`REDIS_PASSWORD`，默认 `opspilot-dev`），在宿主机运行本服务时请设置环境变量 `OPSPILOT_REDIS_PASSWORD` 与之一致，或在配置文件中填写 `redis.password`。

## Seed Data

可直接使用根目录 `docker-compose.yml` 中的 seed：

- 用户名：`admin`
- 密码：`admin123`
- install token：`install-token-demo`

## API Paths

- `GET /healthz`
- `POST /api/v1/auth/login`
- `POST /api/v1/agents/register`
- `GET /api/v1/agents/ws`（鉴权：`X-NexCtl-Agent-Id`、`X-NexCtl-Agent-Secret` 请求头，勿使用 URL query）
- `GET /api/v1/me`
- `GET /api/v1/nodes`
- `GET /api/v1/nodes/{nodeID}`
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
