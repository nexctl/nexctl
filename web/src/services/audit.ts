import { apiRequest } from '@/services/api';
import type { AuditLogItem } from '@/types/audit';
import { mockAuditLogs } from '@/utils/mock';

export async function getAuditLogs() {
  const response = await apiRequest<AuditLogItem[] | { items: AuditLogItem[] }>('/audit/logs', undefined, { items: mockAuditLogs });
  return Array.isArray(response) ? response : response.items;
}
