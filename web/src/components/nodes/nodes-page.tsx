'use client';

import { PageCard, PageHeaderCard, PageShell } from '@/components/layout/page-shell';
import { NodeTable } from '@/components/nodes/node-table';
import { NodesToolbar } from '@/components/nodes/nodes-toolbar';
import type { NodeItem } from '@/types/node';

export function NodesPage({ nodes }: { nodes: NodeItem[] }) {
  return (
    <PageShell>
      <PageHeaderCard titleKey="pages.nodes.title" descriptionKey="pages.nodes.description" />
      <NodesToolbar />
      <PageCard>
        <NodeTable nodes={nodes} />
      </PageCard>
    </PageShell>
  );
}
