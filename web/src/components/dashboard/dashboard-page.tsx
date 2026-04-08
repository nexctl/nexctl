'use client';

import { Alert, Spin } from 'antd';
import { DashboardOverview } from '@/components/dashboard/dashboard-overview';
import { PageHeaderCard, PageShell } from '@/components/layout/page-shell';
import { useNodes } from '@/hooks/use-nodes';

export function DashboardPage() {
  const { nodes, loading, error } = useNodes();

  if (loading) {
    return (
      <PageShell>
        <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      </PageShell>
    );
  }

  if (error) {
    return (
      <PageShell>
        <Alert type="error" title={error} showIcon />
      </PageShell>
    );
  }

  return (
    <PageShell>
      <PageHeaderCard titleKey="pages.dashboard.title" descriptionKey="pages.dashboard.description" />
      <DashboardOverview nodes={nodes} />
    </PageShell>
  );
}
