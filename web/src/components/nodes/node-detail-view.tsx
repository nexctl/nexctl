'use client';

import { Alert, Spin } from 'antd';
import { PageShell } from '@/components/layout/page-shell';
import { NodeDetailPageHeader } from '@/components/nodes/node-detail-page-header';
import { NodeSummaryCards } from '@/components/nodes/node-summary-cards';
import { useNodeDetail } from '@/hooks/use-node-detail';

export function NodeDetailView({ id }: { id: string }) {
  const { detail, loading, error } = useNodeDetail(id);

  if (loading) {
    return (
      <PageShell>
        <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      </PageShell>
    );
  }

  if (error || !detail) {
    return (
      <PageShell>
        <Alert type="error" title={error ?? 'not found'} showIcon />
      </PageShell>
    );
  }

  return (
    <PageShell>
      <NodeDetailPageHeader id={detail.id} name={detail.name} />
      <NodeSummaryCards detail={detail} />
    </PageShell>
  );
}
