'use client';

import { Spin } from 'antd';
import { useRouter } from 'next/navigation';
import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';
import { useAuth } from '@/hooks/use-auth';

/**
 * 首屏在客户端挂载完成前不渲染 Ant Design 组件，避免 antd CSS-in-JS 在 SSR 与浏览器首帧不一致导致 hydration mismatch。
 */
export function AuthGuard({ children }: { children: ReactNode }) {
  const router = useRouter();
  const { user, initialized } = useAuth();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (initialized && !user) {
      router.replace('/login');
    }
  }, [initialized, router, user]);

  if (!mounted) {
    return (
      <div
        style={{ display: 'grid', placeItems: 'center', minHeight: '100vh', background: '#f5f7fa' }}
        suppressHydrationWarning
      />
    );
  }

  if (!initialized || !user) {
    return (
      <div style={{ display: 'grid', placeItems: 'center', minHeight: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return <>{children}</>;
}

