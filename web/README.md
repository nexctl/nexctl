# NexCtl Web Console

## Development

需安装 [pnpm](https://pnpm.io)（推荐通过 Node 自带的 `corepack enable` 启用）。仓库使用 `pnpm-lock.yaml` 锁版本。

```bash
pnpm install
pnpm dev
```

`pnpm dev` 已绑定 `0.0.0.0:3000`，便于局域网访问。

若通过 **局域网 IP**（如 `http://192.168.0.38:3000`）打开控制台，浏览器可能出现 **`/_next/webpack-hmr` WebSocket 失败**：这是 Next 开发模式下的跨站资源保护。请在 `web/.env.local` 中增加本机在该网段下的 **主机名**（逗号分隔多个），然后**重启** `pnpm dev`：

```bash
# 仅主机名，或与下面二选一
NEXT_DEV_ALLOWED_ORIGINS=192.168.0.38
# 也可写完整来源，配置里会解析为 hostname
# NEXT_DEV_ALLOWED_ORIGINS=http://192.168.0.38:3000
```

详见 [allowedDevOrigins](https://nextjs.org/docs/app/api-reference/config/next-config-js/allowedDevOrigins)。

根目录含 `.npmrc`（`node-linker=hoisted`），减轻 Windows 下 Next `standalone` 构建时的符号链接问题；与 Linux/Docker 兼容。

若出现 **`Cannot find module './xxx.js'`**（webpack chunk 缺失），多为 `.next` 缓存与当前构建不一致：先执行 **`pnpm clean`**，再重新 **`pnpm dev`** 或 **`pnpm build`**。

Default address:

```text
http://localhost:3000
```

## Environment

Use `NEXT_PUBLIC_API_BASE_URL` to point to the NexCtl server API, for example:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
```

「添加节点」一键部署命令里的 **控制面地址**（Agent `server_url`）按以下优先级解析：

1. `NEXT_PUBLIC_AGENT_SERVER_URL`（显式指定时最高优先级）
2. 浏览器访问控制台为 **非 localhost 且端口 3000** 时，假定 API 在同主机 **8080** → `http(s)://<当前主机>:8080`
3. `NEXT_PUBLIC_INTERNAL_API_ORIGIN`：由 `next.config.mjs` 根据 **`INTERNAL_API_BASE_URL`**（与 dev 下 `/api/v1` rewrite 同源）在构建时注入；未单独配 `NEXT_PUBLIC_AGENT_SERVER_URL` 时与后端配置对齐
4. 若 `NEXT_PUBLIC_API_BASE_URL` 为绝对 URL，则取其 `origin`
5. `localhost:3000` → `http://127.0.0.1:8080`
6. 否则为当前页面的 `origin`

安装脚本默认从 GitHub `nexctl/nexctl` 的 `master` 拉取；若需指定分支/标签，可设置：

```bash
NEXT_PUBLIC_NEXCTL_INSTALL_REF=master
```

## Default Dev Login

If you use the repository root `docker-compose.yml`, the seeded login is:

- username: `admin`
- password: `admin123`

## Current Structure

- `src/app`: Next.js app router pages
- `src/layouts`: console layout and navigation
- `src/components`: page-level reusable components
- `src/services`: API client and module services
- `src/store`: login state provider
- `src/types`: shared TypeScript models
- `src/utils`: mock data and browser storage helpers

## Notes

- Login state is stored in `localStorage`.
- Pages try real APIs first and fall back to mock data when backend endpoints are not ready.
- Node list and node detail pages are wired to current-state oriented models rather than long-term monitoring views.
