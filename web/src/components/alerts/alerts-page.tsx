'use client';

import { EventTable, RuleTable } from '@/components/alerts/rule-table';
import { PageCard, PageHeaderCard, PageShell } from '@/components/layout/page-shell';
import type { AlertEvent, AlertRule } from '@/types/alert';

export function AlertsPage({ rules, events }: { rules: AlertRule[]; events: AlertEvent[] }) {
  return (
    <PageShell>
      <PageHeaderCard titleKey="pages.alerts.headerTitle" descriptionKey="pages.alerts.headerDescription" />
      <div className="section-grid">
        <PageCard titleKey="pages.alerts.rulesTitle">
          <RuleTable rules={rules} />
        </PageCard>
        <PageCard titleKey="pages.alerts.eventsTitle">
          <EventTable events={events} />
        </PageCard>
      </div>
    </PageShell>
  );
}
