'use client';

import { App, Button, Space } from 'antd';
import { useT } from '@/i18n';

export function NodeToolbar() {
  const t = useT();
  const { message } = App.useApp();

  const onComingSoon = () => {
    message.info(t('nodes.featureComingSoon'));
  };

  return (
    <Space wrap>
      <Button type="primary" onClick={onComingSoon}>
        {t('nodes.toolbarCmd')}
      </Button>
      <Button onClick={onComingSoon}>{t('nodes.toolbarTerminal')}</Button>
      <Button onClick={onComingSoon}>{t('nodes.toolbarFile')}</Button>
      <Button onClick={onComingSoon}>{t('nodes.toolbarUpgrade')}</Button>
    </Space>
  );
}
