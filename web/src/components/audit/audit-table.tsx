'use client';

import { Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMemo } from 'react';
import { useT } from '@/i18n';
import type { AuditLogItem } from '@/types/audit';

export function AuditTable({ items }: { items: AuditLogItem[] }) {
  const t = useT();
  const columns: ColumnsType<AuditLogItem> = useMemo(
    () => [
      { title: t('audit.colUser'), dataIndex: 'actor', key: 'actor' },
      { title: t('audit.colAction'), dataIndex: 'action', key: 'action' },
      { title: t('audit.colResource'), dataIndex: 'resource', key: 'resource' },
      { title: t('audit.colDetail'), dataIndex: 'detail', key: 'detail' },
      { title: t('audit.colTime'), dataIndex: 'created_at', key: 'created_at' },
    ],
    [t],
  );
  return <Table rowKey="id" dataSource={items} columns={columns} />;
}
