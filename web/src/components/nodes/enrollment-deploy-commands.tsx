'use client';

import { CopyOutlined } from '@ant-design/icons';
import { Alert, App, Button, Divider, Input, Space, Typography } from 'antd';
import { useMemo } from 'react';
import { useT } from '@/i18n';
import {
  buildLinuxInstallCommand,
  buildWindowsInstallLines,
  resolveAgentServerUrl,
} from '@/utils/agent-install';

export type EnrollmentDeployCommandsProps = {
  nodeName: string;
  agentId: string;
  agentSecret: string;
  nodeKey: string;
  /** 控制台节点 ID，写入 agent.yaml 的 node_id（可选） */
  nodeId?: number;
};

export function EnrollmentDeployCommands({ nodeName, agentId, agentSecret, nodeKey, nodeId }: EnrollmentDeployCommandsProps) {
  const t = useT();
  const { message } = App.useApp();

  const serverUrl = useMemo(() => resolveAgentServerUrl(), []);

  const linuxCommand = useMemo(
    () => buildLinuxInstallCommand(serverUrl, agentId, agentSecret, nodeKey, nodeId),
    [serverUrl, agentId, agentSecret, nodeKey, nodeId],
  );

  const windowsCommand = useMemo(
    () => buildWindowsInstallLines(serverUrl, agentId, agentSecret, nodeKey, nodeId),
    [serverUrl, agentId, agentSecret, nodeKey, nodeId],
  );

  const copyText = async (text: string, okKey: 'nodes.copied' | 'nodes.deployCopiedCmd') => {
    try {
      await navigator.clipboard.writeText(text);
      message.success(t(okKey));
    } catch {
      message.warning(t('nodes.copyFailed'));
    }
  };

  return (
    <Space orientation="vertical" size="middle" style={{ width: '100%' }}>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
        {t('nodes.credentialHint', { name: nodeName })}
      </Typography.Paragraph>
      <div>
        <Typography.Text type="secondary">{t('nodes.deployServerUrlLabel')}: </Typography.Text>
        <Typography.Text code copyable>
          {serverUrl}
        </Typography.Text>
      </div>
      <Alert type="warning" showIcon title={t('nodes.credentialSecretWarning')} />

      <div>
        <Typography.Text strong>{t('nodes.fieldAgentId')}</Typography.Text>
        <Input.TextArea
          readOnly
          value={agentId}
          autoSize={{ minRows: 1, maxRows: 3 }}
          style={{ fontFamily: 'monospace', marginTop: 8 }}
        />
      </div>
      <div>
        <Typography.Text strong>{t('nodes.fieldAgentSecret')}</Typography.Text>
        <Input.TextArea
          readOnly
          value={agentSecret}
          autoSize={{ minRows: 2, maxRows: 6 }}
          style={{ fontFamily: 'monospace', marginTop: 8 }}
        />
        <Button style={{ marginTop: 8 }} icon={<CopyOutlined />} onClick={() => copyText(agentSecret, 'nodes.copied')}>
          {t('nodes.copyAgentSecret')}
        </Button>
      </div>
      <div>
        <Typography.Text strong>{t('nodes.fieldNodeKey')}</Typography.Text>
        <Input.TextArea
          readOnly
          value={nodeKey}
          autoSize={{ minRows: 2, maxRows: 4 }}
          style={{ fontFamily: 'monospace', marginTop: 8 }}
        />
      </div>

      <Divider>{t('nodes.deploySectionTitle')}</Divider>

      <div>
        <Typography.Text strong>{t('nodes.deployLinuxLabel')}</Typography.Text>
        <Input.TextArea
          readOnly
          value={linuxCommand}
          autoSize={{ minRows: 2, maxRows: 5 }}
          style={{ fontFamily: 'monospace', fontSize: 12, marginTop: 8 }}
        />
        <Button
          style={{ marginTop: 8 }}
          icon={<CopyOutlined />}
          onClick={() => copyText(linuxCommand, 'nodes.deployCopiedCmd')}
        >
          {t('nodes.deployCopyLinux')}
        </Button>
      </div>

      <div>
        <Typography.Text strong>{t('nodes.deployWindowsLabel')}</Typography.Text>
        <Alert
          type="warning"
          showIcon
          style={{ marginTop: 8 }}
          message={t('nodes.deployWindowsAdminHint')}
        />
        <Input.TextArea
          readOnly
          value={windowsCommand}
          autoSize={{ minRows: 3, maxRows: 8 }}
          style={{ fontFamily: 'monospace', fontSize: 12, marginTop: 8 }}
        />
        <Button
          style={{ marginTop: 8 }}
          icon={<CopyOutlined />}
          onClick={() => copyText(windowsCommand, 'nodes.deployCopiedCmd')}
        >
          {t('nodes.deployCopyWindows')}
        </Button>
      </div>
    </Space>
  );
}
