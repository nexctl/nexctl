# NexCtl

NexCtl is a lightweight server fleet monitoring and remote management platform composed of:

- `server`: Go control plane
- `agent`: Go `agentd` + `supervisor`
- `web`: Next.js management console

## Unified Development Contracts

### REST base path

- Base: `http://localhost:8080/api/v1`
- Login: `POST /auth/login`
- Agent register: `POST /agents/register`
- Agent websocket: `GET /agents/ws`（鉴权头 `X-NexCtl-Agent-Id`、`X-NexCtl-Agent-Secret`）
- Nodes: `GET /nodes`, `GET /nodes/{nodeID}`, `POST /nodes/{nodeID}/runtime-state`
- Reserved list APIs: `GET /tasks`, `GET /files`, `GET /upgrades/releases`, `GET /alerts/rules`, `GET /alerts/events`, `GET /audit/logs`

### API envelope

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

### WebSocket envelope

```json
{
  "type": "heartbeat",
  "request_id": "optional-id",
  "timestamp": "2026-04-08T10:30:00Z",
  "payload": {}
}
```

Reserved message types:

- `heartbeat`
- `runtime_state`
- `ack`
- `error`
- `task_dispatch`
- `file_dispatch`
- `upgrade_command`

## Quick Start With Docker Compose

```bash
docker compose up --build
```

After startup:

- Web: `http://localhost:3000`
- Server API: `http://localhost:8080/api/v1`
- MySQL: `127.0.0.1:3306`
- Redis: `127.0.0.1:6379`

Development seed data:

- Console user: `admin`
- Console password: `admin123`
- Install token: `install-token-demo`

## Local Development

1. Start dependencies:

```bash
docker compose up mysql redis
```

   If Redis is started via this Compose file, it uses password `opspilot-dev` by default. Export `OPSPILOT_REDIS_PASSWORD=opspilot-dev` (or match your `REDIS_PASSWORD`) before running the server locally.

2. Start server:

```bash
cd server
go run ./cmd/server -config configs/config.example.yaml
```

3. Start web:

```bash
cd web
pnpm install
pnpm dev
```

4. Start agent:

```bash
cd agent
go run ./cmd/agentd -config configs/agent.example.yaml
```
