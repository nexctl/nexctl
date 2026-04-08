'use client';

import { Button, Form, Input, Select, Space } from 'antd';
import { useT } from '@/i18n';

export function TaskFilter() {
  const t = useT();
  return (
    <Form layout="inline">
      <Form.Item name="status">
        <Select
          style={{ width: 160 }}
          placeholder={t('tasks.filterStatus')}
          options={[
            { label: t('tasks.filterAllStatus'), value: '' },
            { label: 'pending', value: 'pending' },
            { label: 'running', value: 'running' },
            { label: 'success', value: 'success' },
            { label: 'failed', value: 'failed' },
          ]}
        />
      </Form.Item>
      <Form.Item name="keyword">
        <Input placeholder={t('tasks.filterKeyword')} style={{ width: 220 }} />
      </Form.Item>
      <Space>
        <Button type="primary">{t('common.query')}</Button>
        <Button>{t('common.reset')}</Button>
      </Space>
    </Form>
  );
}
