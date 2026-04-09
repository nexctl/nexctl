'use client';

import { Button, Card, Space, Typography } from 'antd';
import { useT } from '@/i18n';

export function TasksPageHeader({ onNewTask }: { onNewTask?: () => void }) {
  const t = useT();
  return (
    <Card className="page-card">
      <Space style={{ width: '100%', justifyContent: 'space-between' }} align="start">
        <div>
          <Typography.Title level={3} style={{ marginBottom: 0 }}>
            {t('tasks.headerTitle')}
          </Typography.Title>
          <Typography.Text type="secondary">{t('tasks.headerDesc')}</Typography.Text>
        </div>
        <Button type="primary" onClick={() => onNewTask?.()}>
          {t('tasks.newTask')}
        </Button>
      </Space>
    </Card>
  );
}
