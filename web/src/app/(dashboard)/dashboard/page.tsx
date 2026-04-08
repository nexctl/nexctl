import { DashboardPage } from '@/components/dashboard/dashboard-page';
import { getNodes } from '@/services/node';

export default async function DashboardRoutePage() {
  const nodes = await getNodes();
  return <DashboardPage nodes={nodes} />;
}
