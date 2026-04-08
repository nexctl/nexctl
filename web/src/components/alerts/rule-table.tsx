'use client';

import { Switch, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMemo } from 'react';
import { useT } from '@/i18n';
import type { AlertEvent, AlertRule } from '@/types/alert';

export function RuleTable({ rules }: { rules: AlertRule[] }) {
  const t = useT();
  const columns: ColumnsType<AlertRule> = useMemo(
    () => [
      { title: t('alerts.colRuleName'), dataIndex: 'name', key: 'name' },
      { title: t('alerts.colType'), dataIndex: 'type', key: 'type' },
      { title: t('alerts.colTarget'), dataIndex: 'target', key: 'target' },
      {
        title: t('alerts.colEnabled'),
        dataIndex: 'enabled',
        key: 'enabled',
        render: (value: boolean) => <Switch checked={value} />,
      },
    ],
    [t],
  );
  return <Table rowKey="id" pagination={false} dataSource={rules} columns={columns} />;
}

export function EventTable({ events }: { events: AlertEvent[] }) {
  const t = useT();
  const columns: ColumnsType<AlertEvent> = useMemo(
    () => [
      {
        title: t('alerts.colLevel'),
        dataIndex: 'severity',
        key: 'severity',
        render: (value: string) => <Tag color={value === 'critical' ? 'red' : 'orange'}>{value}</Tag>,
      },
      { title: t('alerts.colNode'), dataIndex: 'node_name', key: 'node_name' },
      { title: t('alerts.colSummary'), dataIndex: 'summary', key: 'summary' },
      { title: t('alerts.colStatus'), dataIndex: 'status', key: 'status' },
      { title: t('alerts.colTime'), dataIndex: 'created_at', key: 'created_at' },
    ],
    [t],
  );
  return <Table rowKey="id" dataSource={events} columns={columns} />;
}
