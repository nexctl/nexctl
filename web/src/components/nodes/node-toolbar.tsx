'use client';

import { Button, Space } from 'antd';
import { useT } from '@/i18n';

export function NodeToolbar() {
  const t = useT();
  return (
    <Space wrap>
      <Button type="primary">{t('nodes.toolbarCmd')}</Button>
      <Button>{t('nodes.toolbarTerminal')}</Button>
      <Button>{t('nodes.toolbarFile')}</Button>
      <Button>{t('nodes.toolbarUpgrade')}</Button>
    </Space>
  );
}
