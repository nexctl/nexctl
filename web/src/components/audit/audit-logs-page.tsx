'use client';

import { AuditFilter } from '@/components/audit/audit-filter';
import { AuditTable } from '@/components/audit/audit-table';
import { PageCard, PageHeaderCard, PageShell } from '@/components/layout/page-shell';
import type { AuditLogItem } from '@/types/audit';

export function AuditLogsPage({ logs }: { logs: AuditLogItem[] }) {
  return (
    <PageShell>
      <PageHeaderCard titleKey="pages.audit.title" descriptionKey="pages.audit.description" />
      <PageCard>
        <AuditFilter />
      </PageCard>
      <PageCard>
        <AuditTable items={logs} />
      </PageCard>
    </PageShell>
  );
}
