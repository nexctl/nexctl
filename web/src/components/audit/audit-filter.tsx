'use client';

import { Button, Form, Input, Space } from 'antd';
import { useT } from '@/i18n';

export function AuditFilter() {
  const t = useT();
  return (
    <Form layout="inline">
      <Form.Item name="actor">
        <Input placeholder={t('audit.filterUser')} style={{ width: 200 }} />
      </Form.Item>
      <Form.Item name="action">
        <Input placeholder={t('audit.filterAction')} style={{ width: 220 }} />
      </Form.Item>
      <Space>
        <Button type="primary">{t('common.query')}</Button>
        <Button>{t('common.reset')}</Button>
      </Space>
    </Form>
  );
}
