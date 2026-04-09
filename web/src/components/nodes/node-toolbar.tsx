'use client';

import { App, Button, Space } from 'antd';
import { useState } from 'react';
import { NodeFileManagerModal } from '@/components/nodes/node-file-manager-modal';
import { NodeTerminalModal } from '@/components/nodes/node-terminal-modal';
import { useT } from '@/i18n';
import { triggerAgentUpgrade } from '@/services/node';

export function NodeToolbar({ nodeId, nodeName, platform }: { nodeId: number; nodeName: string; platform?: string }) {
  const t = useT();
  const { message } = App.useApp();
  const [terminalOpen, setTerminalOpen] = useState(false);
  const [fileOpen, setFileOpen] = useState(false);
  const [upgradeLoading, setUpgradeLoading] = useState(false);

  const onComingSoon = () => {
    message.info(t('nodes.featureComingSoon'));
  };

  const onUpgradeAgent = async () => {
    setUpgradeLoading(true);
    try {
      await triggerAgentUpgrade(nodeId);
      message.success(t('nodes.toolbarUpgradeQueued'));
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('nodes.toolbarUpgradeFailed'));
    } finally {
      setUpgradeLoading(false);
    }
  };

  return (
    <>
      <Space wrap>
        <Button type="primary" onClick={onComingSoon}>
          {t('nodes.toolbarCmd')}
        </Button>
        <Button onClick={() => setTerminalOpen(true)}>{t('nodes.toolbarTerminal')}</Button>
        <Button onClick={() => setFileOpen(true)}>{t('nodes.toolbarFile')}</Button>
        <Button loading={upgradeLoading} onClick={onUpgradeAgent}>
          {t('nodes.toolbarUpgrade')}
        </Button>
      </Space>
      <NodeTerminalModal
        open={terminalOpen}
        onClose={() => setTerminalOpen(false)}
        nodeId={nodeId}
        nodeName={nodeName}
      />
      <NodeFileManagerModal
        open={fileOpen}
        onClose={() => setFileOpen(false)}
        nodeId={nodeId}
        nodeName={nodeName}
        platform={platform}
      />
    </>
  );
}
