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
