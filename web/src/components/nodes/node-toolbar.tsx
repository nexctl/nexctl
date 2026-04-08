'use client';

import { App, Button, Space } from 'antd';
import { useState } from 'react';
import { NodeTerminalModal } from '@/components/nodes/node-terminal-modal';
import { useT } from '@/i18n';

export function NodeToolbar({ nodeId, nodeName }: { nodeId: number; nodeName: string }) {
  const t = useT();
  const { message } = App.useApp();
  const [terminalOpen, setTerminalOpen] = useState(false);

  const onComingSoon = () => {
    message.info(t('nodes.featureComingSoon'));
  };

  return (
    <>
      <Space wrap>
        <Button type="primary" onClick={onComingSoon}>
          {t('nodes.toolbarCmd')}
        </Button>
        <Button onClick={() => setTerminalOpen(true)}>{t('nodes.toolbarTerminal')}</Button>
        <Button onClick={onComingSoon}>{t('nodes.toolbarFile')}</Button>
        <Button onClick={onComingSoon}>{t('nodes.toolbarUpgrade')}</Button>
      </Space>
      <NodeTerminalModal
        open={terminalOpen}
        onClose={() => setTerminalOpen(false)}
        nodeId={nodeId}
        nodeName={nodeName}
      />
    </>
  );
}
