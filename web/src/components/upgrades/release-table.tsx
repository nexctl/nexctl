'use client';

import { Button, Space, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMemo } from 'react';
import { useT } from '@/i18n';
import type { ReleaseItem } from '@/types/upgrade';

export function ReleaseTable({ releases }: { releases: ReleaseItem[] }) {
  const t = useT();

  const columns: ColumnsType<ReleaseItem> = useMemo(
    () => [
      { title: t('upgrades.colVersion'), dataIndex: 'version', key: 'version' },
      { title: t('upgrades.colChannel'), dataIndex: 'channel', key: 'channel' },
      { title: t('upgrades.colNotes'), dataIndex: 'notes', key: 'notes' },
      {
        title: t('upgrades.colPublishStatus'),
        dataIndex: 'rollout_status',
        key: 'rollout_status',
        render: (value: string) => <Tag color={value === 'rolling' ? 'blue' : 'green'}>{value}</Tag>,
      },
      { title: t('upgrades.colCreated'), dataIndex: 'created_at', key: 'created_at' },
      {
        title: t('upgrades.colAction'),
        key: 'actions',
        render: () => (
          <Space>
            <Button size="small" type="primary">
              {t('upgrades.canaryDeploy')}
            </Button>
            <Button size="small">{t('upgrades.viewNodes')}</Button>
          </Space>
        ),
      },
    ],
    [t],
  );

  return <Table rowKey="id" dataSource={releases} columns={columns} />;
}
