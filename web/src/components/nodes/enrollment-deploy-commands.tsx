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
  token: string;
  expiresAt?: string;
};

export function EnrollmentDeployCommands({ nodeName, token, expiresAt }: EnrollmentDeployCommandsProps) {
  const t = useT();
  const { message } = App.useApp();

  const serverUrl = useMemo(() => resolveAgentServerUrl(), [token]);

  const linuxCommand = useMemo(
    () => buildLinuxInstallCommand(serverUrl, token),
    [serverUrl, token],
  );

  const windowsCommand = useMemo(
    () => buildWindowsInstallLines(serverUrl, token),
    [serverUrl, token],
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
        {t('nodes.tokenHint', { name: nodeName })}
      </Typography.Paragraph>
      <div>
        <Typography.Text type="secondary">{t('nodes.deployServerUrlLabel')}: </Typography.Text>
        <Typography.Text code copyable>
          {serverUrl}
        </Typography.Text>
      </div>
      <Alert type="info" showIcon title={t('nodes.deployServerHint')} />

      {expiresAt && (
        <Typography.Text type="secondary">
          {t('nodes.expiresAt')}
          {expiresAt}
        </Typography.Text>
      )}

      <div>
        <Typography.Text strong>{t('nodes.copyToken')}</Typography.Text>
        <Input.TextArea
          readOnly
          value={token}
          autoSize={{ minRows: 3, maxRows: 6 }}
          style={{ fontFamily: 'monospace', marginTop: 8 }}
        />
        <Button style={{ marginTop: 8 }} icon={<CopyOutlined />} onClick={() => copyText(token, 'nodes.copied')}>
          {t('nodes.copyToken')}
        </Button>
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
