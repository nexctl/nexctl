/** @type {import('next').NextConfig} */
const internalApiBase = (process.env.INTERNAL_API_BASE_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');

const nextConfig = {
  ...(process.env.NODE_ENV === 'production' ? { output: 'standalone' } : {}),
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

