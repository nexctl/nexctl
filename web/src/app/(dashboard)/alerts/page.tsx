import { AlertsPage } from '@/components/alerts/alerts-page';
import { getAlertEvents, getAlertRules } from '@/services/alert';

export default async function AlertsRoutePage() {
  const [rules, events] = await Promise.all([getAlertRules(), getAlertEvents()]);
  return <AlertsPage rules={rules} events={events} />;
}
