'use client';

import { Button, Card, Space, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useRouter } from 'next/navigation';
import { useMemo, useState } from 'react';
import { PageShell } from '@/components/layout/page-shell';
import { CreateScheduleModal } from '@/components/task-schedules/create-schedule-modal';
import { useT } from '@/i18n';
import type { TaskScheduleItem } from '@/services/task';

export function TaskSchedulesPage({ schedules }: { schedules: TaskScheduleItem[] }) {
  const t = useT();
  const router = useRouter();
  const [open, setOpen] = useState(false);

  const columns: ColumnsType<TaskScheduleItem> = useMemo(
    () => [
      { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
      { title: t('taskSchedules.colName'), dataIndex: 'name', key: 'name', ellipsis: true },
      { title: t('taskSchedules.colCron'), dataIndex: 'cron_expr', key: 'cron_expr', width: 140 },
      { title: t('tasks.colType'), dataIndex: 'task_type', key: 'task_type', width: 120 },
      { title: t('tasks.colScope'), dataIndex: 'scope', key: 'scope', ellipsis: true },
      {
        title: t('taskSchedules.colEnabled'),
        dataIndex: 'enabled',
        key: 'enabled',
        width: 96,
        render: (v: boolean) => (
          <Tag color={v ? 'green' : 'default'}>{v ? t('taskSchedules.enabledOn') : t('taskSchedules.enabledOff')}</Tag>
        ),
      },
      {
        title: t('taskSchedules.colNextRun'),
        dataIndex: 'next_run_at',
        key: 'next_run_at',
        width: 200,
        render: (v: string) => (v ? new Date(v).toLocaleString() : '—'),
      },
      {
        title: t('tasks.newTaskDetail'),
        dataIndex: 'detail',
        key: 'detail',
        ellipsis: true,
      },
    ],
    [t],
  );

  return (
    <PageShell>
      <Card className="page-card">
        <Space style={{ width: '100%', justifyContent: 'space-between' }} align="start">
          <div>
            <Typography.Title level={3} style={{ marginBottom: 0 }}>
              {t('taskSchedules.headerTitle')}
            </Typography.Title>
            <Typography.Text type="secondary">{t('taskSchedules.headerDesc')}</Typography.Text>
          </div>
          <Button type="primary" onClick={() => setOpen(true)}>
            {t('taskSchedules.newSchedule')}
          </Button>
        </Space>
      </Card>
      <Card className="page-card">
        <Table rowKey="id" dataSource={schedules} columns={columns} scroll={{ x: 1100 }} />
      </Card>
      <CreateScheduleModal
        open={open}
        onClose={() => setOpen(false)}
        onCreated={() => router.refresh()}
      />
    </PageShell>
  );
}
