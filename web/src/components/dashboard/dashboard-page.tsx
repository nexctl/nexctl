'use client';

import { DashboardOverview } from '@/components/dashboard/dashboard-overview';
import { PageHeaderCard, PageShell } from '@/components/layout/page-shell';
import type { NodeItem } from '@/types/node';

export function DashboardPage({ nodes }: { nodes: NodeItem[] }) {
  return (
    <PageShell>
      <PageHeaderCard titleKey="pages.dashboard.title" descriptionKey="pages.dashboard.description" />
      <DashboardOverview nodes={nodes} />
    </PageShell>
  );
}
