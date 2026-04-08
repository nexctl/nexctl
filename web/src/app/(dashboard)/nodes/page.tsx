import { NodesPage } from '@/components/nodes/nodes-page';
import { getNodes } from '@/services/node';

export default async function NodesRoutePage() {
  const nodes = await getNodes();
  return <NodesPage nodes={nodes} />;
}
