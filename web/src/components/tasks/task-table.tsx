'use client';

import { Button, Drawer, Progress, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMemo, useState } from 'react';
import { useT } from '@/i18n';
import type { TaskItem } from '@/types/task';

export function TaskTable({ tasks }: { tasks: TaskItem[] }) {
  const t = useT();
  const [activeTask, setActiveTask] = useState<TaskItem | null>(null);

  const columns: ColumnsType<TaskItem> = useMemo(
    () => [
      { title: t('tasks.colId'), dataIndex: 'id', key: 'id' },
      { title: t('tasks.colType'), dataIndex: 'type', key: 'type' },
      { title: t('tasks.colScope'), dataIndex: 'scope', key: 'scope' },
      { title: t('tasks.colOperator'), dataIndex: 'operator', key: 'operator' },
      {
        title: t('tasks.colStatus'),
        dataIndex: 'status',
        key: 'status',
        render: (value: string) => (
          <Tag color={value === 'success' ? 'green' : value === 'running' ? 'blue' : 'default'}>{value}</Tag>
        ),
      },
      {
        title: t('tasks.colProgress'),
        key: 'progress',
        render: (_, row: TaskItem) => <Progress percent={row.progress} size="small" />,
      },
      { title: t('tasks.colCreated'), dataIndex: 'created_at', key: 'created_at' },
      {
        title: t('tasks.colAction'),
        key: 'action',
        render: (_, row: TaskItem) => (
          <Button onClick={() => setActiveTask(row)}>{t('nodes.detail')}</Button>
        ),
      },
    ],
    [t],
  );

  return (
    <>
      <Table rowKey="id" dataSource={tasks} columns={columns} />
      <Drawer
        open={Boolean(activeTask)}
        onClose={() => setActiveTask(null)}
        title={t('tasks.detailTitle', { id: activeTask?.id ?? '' })}
        size={520}
      >
        {activeTask && (
          <>
            <Typography.Paragraph>
              {t('tasks.detailType')}
              {activeTask.type}
            </Typography.Paragraph>
            <Typography.Paragraph>
              {t('tasks.detailStatus')}
              {activeTask.status}
            </Typography.Paragraph>
            <Typography.Paragraph>
              {t('tasks.detailBody')}
              {activeTask.detail}
            </Typography.Paragraph>
            <Progress percent={activeTask.progress} />
          </>
        )}
      </Drawer>
    </>
  );
}
