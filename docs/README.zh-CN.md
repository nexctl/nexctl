# NexCtl 中文文档

## 项目概览

NexCtl 是一个轻量服务器集群监控与管理系统，当前 monorepo 包含三个主要子项目：

- `server`：Go 控制面服务端
- `agent`：Go 编写的 `agentd` 与 `supervisor`
- `web`：Next.js + React + TypeScript 管理后台

当前阶段重点是：

- 节点注册与长期凭证签发
- Agent WebSocket 长连接
- 当前状态采集与短期状态展示
- 基础后台登录、节点列表与节点详情
- 为 task / file / upgrade / alert / audit 模块预留统一接口

不包含：

- 长期时序指标存储
- 完整任务编排实现
- 完整升级下载替换流程

## 目录说明

### 根目录

- `server/`：控制面服务
- `agent/`：Agent 与 Supervisor
- `web/`：管理后台
- `deploy/`：安装与开发环境附加文件
- `docs/`：文档目录
- `docker-compose.yml`：开发联调用环境

### server

- `cmd/server`：启动入口
- `internal/api`：handler、middleware、router
- `internal/auth`：登录逻辑
- `internal/node`：节点注册、列表、详情 DTO 与服务
- `internal/runtime`：当前状态更新逻辑
- `internal/ws`：Agent WebSocket 协议处理
- `internal/repository`：MySQL / Redis 数据访问
- `internal/task` / `internal/filemgr` / `internal/upgrade` / `internal/alert` / `internal/audit`：后续模块统一接口预留
- `migrations`：数据库初始化脚本

### agent

- `cmd/agentd`：主 Agent 入口
- `cmd/supervisor`：守护进程入口
- `internal/app`：agentd / supervisor 启动编排
- `internal/config`：配置加载
- `internal/collector`：运行态与平台信息采集
- `internal/transport/httpclient`：注册 HTTP 客户端
- `internal/transport/wsclient`：WebSocket 客户端
- `internal/platform`：平台抽象与平台实现
- `internal/upgrader`：升级目录和流程预留
- `internal/store`：本地凭证与节点标识存储

### web

- `src/app`：Next.js App Router 页面
- `src/components`：页面拆分组件
- `src/layouts`：后台布局
- `src/services`：统一 API 调用层
- `src/types`：类型定义
- `src/store`：登录态管理
- `src/utils`：mock 数据和浏览器存储工具

## 统一接口约定

### REST 基础路径

所有服务端接口统一以：

```text
http://localhost:8080/api/v1
```

为前缀。

### API 返回结构

统一使用：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

### 已统一的核心接口

- `POST /api/v1/auth/login`
- `POST /api/v1/agents/register`
- `GET /api/v1/agents/ws`
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

### 控制台 RBAC（JWT 角色码）

- `admin` / `super_admin` / `root`：全部受保护接口。
- `viewer` / `readonly`：仅 `nodes:read` 与 `modules:read`（可读节点与模块占位列表），**不含** `nodes:write`（不能通过控制台接口上报节点 runtime）。
- 其它角色码：默认拒绝需要权限的路由。

权限与路由对应关系：

| 权限 | 路由 |
|------|------|
| `nodes:read` | `GET /nodes`、`GET /nodes/{nodeID}` |
| `nodes:write` | `POST /nodes/{nodeID}/runtime-state` |
| `modules:read` | `GET /tasks`、`/files`、`/upgrades/releases`、`/alerts/*`、`/audit/logs` |

登录与 Agent 注册接口带有按 IP 的速率限制（减轻暴力尝试）。

### WebSocket 消息协议

统一包结构：

```json
{
  "type": "runtime_state",
  "request_id": "req-123",
  "timestamp": "2026-04-08T10:30:00Z",
  "payload": {}
}
```

当前已实现消息类型：

- `heartbeat`
- `runtime_state`
- `ack`
- `error`

已预留消息类型：

- `task_dispatch`
- `file_dispatch`
- `upgrade_command`

### Agent 注册返回长期凭证

注册成功后，服务端返回：

- `node_id`
- `agent_id`
- `agent_secret`
- `ws_url`

Agent 会持久化到本地凭证文件，后续重启优先复用，不重复注册。

### Agent WebSocket 鉴权

连接 `GET /api/v1/agents/ws` 时必须在 **WebSocket HTTP 升级请求** 中设置请求头（不要使用 URL query，避免进入访问日志、Referer 等）：

- `X-NexCtl-Agent-Id`
- `X-NexCtl-Agent-Secret`

带 `Origin` 头的浏览器客户端，其 Origin 须匹配服务端配置中的 `app.websocket_allowed_origins`（可用环境变量 `OPSPILOT_WEBSOCKET_ALLOWED_ORIGINS`，逗号分隔）。无 `Origin` 的非浏览器客户端（如本仓库自带 Agent）不受该列表限制。

## Docker 开发环境

仓库根目录已经提供：

- `docker-compose.yml`
- `server/Dockerfile.dev`
- `web/Dockerfile.dev`

### 启动方式

```bash
docker compose up --build
```

启动后默认地址：

- Web：`http://localhost:3000`
- Server API：`http://localhost:8080/api/v1`
- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6379`（Compose 默认启用口令，见下）
- 控制面 HTTP：`127.0.0.1:8080`（仅本机回环，避免在局域网直接暴露 API）

Compose 中 Redis 默认 `REDIS_PASSWORD=opspilot-dev`，server 服务已通过 `OPSPILOT_REDIS_PASSWORD` 注入。若你**只起 `mysql redis` 后在宿主机运行 server**，请设置相同环境变量，或在 `config.yaml` 的 `redis.password` 中填写该值。

**数据库结构变更后**（例如更新了 `migrations/0001_init.sql`），需删除对应 Docker volume 或重建库后再执行初始化脚本。

### 默认开发数据

- 用户名：`admin`
- 密码：`admin123`
- install token：`install-token-demo`

## 本地开发启动

### 1. 启动依赖

```bash
docker compose up mysql redis
```

若使用上述 Compose 中的 Redis（默认带口令），在启动 server 前设置（PowerShell 示例）：

```powershell
$env:OPSPILOT_REDIS_PASSWORD = "opspilot-dev"
```

### 2. 启动 server

```bash
cd server
go run ./cmd/server -config configs/config.example.yaml
```

### 3. 启动 web

```bash
cd web
pnpm install
pnpm dev
```

### 4. 启动 agentd

先确认 `agent/configs/agent.example.yaml` 中：

- `server_url` 指向 `http://localhost:8080`
- `register_path` 为 `/api/v1/agents/register`
- `install_token` 为有效 token

然后运行：

```bash
cd agent
go run ./cmd/agentd -config configs/agent.example.yaml
```

### 5. 启动 supervisor

```bash
cd agent
go run ./cmd/supervisor -config configs/supervisor.example.yaml
```

## 当前联调状态

### 已完成

- `server` 可编译
- `agent` 可编译
- `web` 可构建
- API 路径已统一
- WebSocket 包结构已统一
- 节点注册 DTO、节点列表 DTO 已统一
- task / file / upgrade / alert / audit 已补统一占位接口

### 当前限制

- `task` / `file` / `upgrade` / `alert` / `audit` 仍是最小返回实现
- 节点详情中的服务、任务、告警和短期小图数据目前仍以占位或 mock 为主
- Agent 升级与回滚目录结构已预留，但完整升级替换流程尚未完成

## 建议下一步

- 打通真实 `task` 列表和执行详情
- 打通 `file` 上传与分发记录
- 为 `node_labels` 增加真实查询与返回
- 在 `server` 中补齐 `short_term_metrics` 的 Redis 读取接口
- 在 `web` 中把节点详情页切换到真实短期状态折线数据
