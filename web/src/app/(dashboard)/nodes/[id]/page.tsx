import { PageShell } from '@/components/layout/page-shell';
import { NodeDetailPageHeader } from '@/components/nodes/node-detail-page-header';
import { NodeSummaryCards } from '@/components/nodes/node-summary-cards';
import { getNodeDetail } from '@/services/node';

export default async function NodeDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const detail = await getNodeDetail(id);

  return (
    <PageShell>
      <NodeDetailPageHeader id={detail.id} name={detail.name} />
      <NodeSummaryCards detail={detail} />
    </PageShell>
  );
}
