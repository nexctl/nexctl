export interface AuditLogItem {
  id: number;
  actor: string;
  action: string;
  resource: string;
  detail: string;
  created_at: string;
}

