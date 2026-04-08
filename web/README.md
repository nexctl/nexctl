# NexCtl Web Console

## Development

需安装 [pnpm](https://pnpm.io)（推荐通过 Node 自带的 `corepack enable` 启用）。仓库使用 `pnpm-lock.yaml` 锁版本。

```bash
pnpm install
pnpm dev
```

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

1. `NEXT_PUBLIC_AGENT_SERVER_URL`（例如 `https://console.example.com` 或 `http://10.0.0.1:8080`）
2. 若 `NEXT_PUBLIC_API_BASE_URL` 为绝对 URL，则取其 `origin`
3. 本地开发常见 `localhost:3000` 时默认 `http://127.0.0.1:8080`
4. 否则为当前页面的 `origin`

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
