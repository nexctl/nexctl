'use client';

import { Alert, Spin } from 'antd';
import { PageCard, PageHeaderCard, PageShell } from '@/components/layout/page-shell';
import { NodeTable } from '@/components/nodes/node-table';
import { NodesToolbar } from '@/components/nodes/nodes-toolbar';
import { useNodes } from '@/hooks/use-nodes';

export function NodesPage() {
  const { nodes, loading, error, refetch } = useNodes();

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
      <PageHeaderCard titleKey="pages.nodes.title" descriptionKey="pages.nodes.description" />
      <NodesToolbar onNodeCreated={refetch} />
      <PageCard>
        <NodeTable nodes={nodes} onAfterDelete={refetch} />
      </PageCard>
    </PageShell>
  );
}
