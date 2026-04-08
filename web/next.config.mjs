/** @type {import('next').NextConfig} */
const internalApiBase = (process.env.INTERNAL_API_BASE_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');

/** 解析 NEXT_DEV_ALLOWED_ORIGINS：逗号分隔，可为纯主机名或含协议的完整 URL（如 http://192.168.0.38:3000） */
function parseAllowedDevOrigins(raw) {
  if (!raw || typeof raw !== 'string') return [];
  return raw
    .split(',')
    .map((s) => {
      const t = s.trim();
      if (!t) return '';
      try {
        if (t.includes('://')) {
          return new URL(t).hostname;
        }
      } catch {
        return t;
      }
      return t;
    })
    .filter(Boolean);
}

const isProd = process.env.NODE_ENV === 'production';
const allowedDevOrigins = !isProd ? parseAllowedDevOrigins(process.env.NEXT_DEV_ALLOWED_ORIGINS) : [];

const nextConfig = {
  ...(!isProd && allowedDevOrigins.length > 0 ? { allowedDevOrigins } : {}),
  ...(isProd ? { output: 'standalone' } : {}),
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: `${internalApiBase}/api/v1/:path*`,
      },
    ];
  },
};

export default nextConfig;
