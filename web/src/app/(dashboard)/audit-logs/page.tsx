import { AuditLogsPage } from '@/components/audit/audit-logs-page';
import { getAuditLogs } from '@/services/audit';

export default async function AuditLogsRoutePage() {
  const logs = await getAuditLogs();
  return <AuditLogsPage logs={logs} />;
}
