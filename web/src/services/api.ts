import type { ApiEnvelope } from '@/types/common';
import { readAuthStorage } from '@/utils/storage';

/**
 * 浏览器：走同源 `/api/v1`（由 Next rewrites 转到后端）。
 * 服务端组件（RSC）：必须请求绝对地址；直连 INTERNAL_API_BASE_URL，避免对相对 URL 的 fetch 在 Node 中解析失败或误指向错误主机。
 */
function resolveApiBaseUrl(): string {
  const fromPublic = (process.env.NEXT_PUBLIC_API_BASE_URL ?? '/api/v1').replace(/\/$/, '');
  if (typeof window !== 'undefined') {
    return fromPublic;
  }
  const internal = process.env.INTERNAL_API_BASE_URL?.replace(/\/$/, '');
  if (internal) {
    return `${internal}/api/v1`;
  }
  if (fromPublic.startsWith('http://') || fromPublic.startsWith('https://')) {
    return fromPublic;
  }
  return 'http://127.0.0.1:8080/api/v1';
}

export async function apiRequest<T>(path: string, init?: RequestInit, fallback?: T): Promise<T> {
  const auth = readAuthStorage();
  const base = resolveApiBaseUrl();

  try {
    const response = await fetch(`${base}${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(auth?.token ? { Authorization: `Bearer ${auth.token}` } : {}),
        ...(init?.headers ?? {}),
      },
      cache: 'no-store',
    });

    const text = await response.text();
    let envelope: ApiEnvelope<T> | null = null;
    if (text) {
      try {
        envelope = JSON.parse(text) as ApiEnvelope<T>;
      } catch {
        envelope = null;
      }
    }

    if (!response.ok) {
      const msg = envelope?.message?.trim() || `HTTP ${response.status}`;
      throw new Error(msg);
    }

    if (!envelope) {
      throw new Error(`HTTP ${response.status}: empty body`);
    }
    if (envelope.code !== 0) {
      throw new Error(envelope.message);
    }
    return envelope.data;
  } catch (error) {
    if (fallback !== undefined) {
      return fallback;
    }
    throw error;
  }
}
