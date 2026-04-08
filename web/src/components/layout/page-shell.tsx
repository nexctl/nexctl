'use client';

import { Card, Space, Typography } from 'antd';
import type { ReactNode } from 'react';
import { useT } from '@/i18n';

export function PageShell({ children }: { children: ReactNode }) {
  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      {children}
    </Space>
  );
}

export function PageHeaderCard({ titleKey, descriptionKey }: { titleKey: string; descriptionKey: string }) {
  const t = useT();
  return (
    <Card className="page-card">
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        {t(titleKey)}
      </Typography.Title>
      <Typography.Text type="secondary">{t(descriptionKey)}</Typography.Text>
    </Card>
  );
}

export function PageCard({
  titleKey,
  children,
  className,
}: {
  titleKey?: string;
  children: ReactNode;
  className?: string;
}) {
  const t = useT();
  const title = titleKey ? t(titleKey) : undefined;
  return (
    <Card className={className ?? 'page-card'} title={title}>
      {children}
    </Card>
  );
}
