'use client';

import { Button, Card, Space, Typography } from 'antd';
import { useT } from '@/i18n';

export function FilesPageHeader() {
  const t = useT();
  return (
    <Card className="page-card">
      <Space style={{ width: '100%', justifyContent: 'space-between' }} align="start">
        <div>
          <Typography.Title level={3} style={{ marginBottom: 0 }}>
            {t('files.headerTitle')}
          </Typography.Title>
          <Typography.Text type="secondary">{t('files.headerDesc')}</Typography.Text>
        </div>
        <Space>
          <Button type="primary">{t('files.upload')}</Button>
          <Button>{t('files.distribution')}</Button>
        </Space>
      </Space>
    </Card>
  );
}
