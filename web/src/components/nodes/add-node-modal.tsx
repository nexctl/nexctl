'use client';

import { CopyOutlined } from '@ant-design/icons';
import { App, Alert, Button, Divider, Form, Input, InputNumber, Modal, Space, Typography } from 'antd';
import { useMemo, useState } from 'react';
import { useT } from '@/i18n';
import { createPendingNode } from '@/services/node';
import {
  buildLinuxInstallCommand,
  buildWindowsInstallLines,
  resolveAgentServerUrl,
} from '@/utils/agent-install';

type Props = {
  open: boolean;
  onClose: () => void;
  onCreated?: () => void;
};

export function AddNodeModal({ open, onClose, onCreated }: Props) {
  const { message } = App.useApp();
  const t = useT();
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);
  const [tokenResult, setTokenResult] = useState<{
    token: string;
    expiresAt?: string;
    name: string;
  } | null>(null);

  const serverUrl = useMemo(() => resolveAgentServerUrl(), [tokenResult]);

  const linuxCommand = useMemo(
    () => (tokenResult ? buildLinuxInstallCommand(serverUrl, tokenResult.token) : ''),
    [tokenResult, serverUrl],
  );

  const windowsCommand = useMemo(
    () => (tokenResult ? buildWindowsInstallLines(serverUrl, tokenResult.token) : ''),
    [tokenResult, serverUrl],
  );

  const handleClose = () => {
    setTokenResult(null);
    form.resetFields();
    onClose();
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    setSubmitting(true);
    try {
      const res = await createPendingNode({
        name: values.name.trim(),
        expires_in_hours: values.expires_in_hours ?? undefined,
      });
      setTokenResult({
        token: res.enrollment_token,
        expiresAt: res.enrollment_expires_at,
        name: res.name,
      });
      message.success(t('nodes.createSuccess'));
      onCreated?.();
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('nodes.createFailed'));
    } finally {
      setSubmitting(false);
    }
  };

  const copyText = async (text: string, okKey: 'nodes.copied' | 'nodes.deployCopiedCmd') => {
    try {
      await navigator.clipboard.writeText(text);
      message.success(t(okKey));
    } catch {
      message.warning(t('nodes.copyFailed'));
    }
  };

  return (
    <Modal
      title={tokenResult ? t('nodes.modalTokenTitle') : t('nodes.modalAddTitle')}
      open={open}
      onCancel={handleClose}
      footer={
        tokenResult ? (
          <Button type="primary" onClick={handleClose}>
            {t('common.close')}
          </Button>
        ) : (
          <Space>
            <Button onClick={handleClose}>{t('common.cancel')}</Button>
            <Button type="primary" loading={submitting} onClick={handleSubmit}>
              {t('nodes.createToken')}
            </Button>
          </Space>
        )
      }
      destroyOnHidden
      width={720}
    >
      {tokenResult ? (
        <Space orientation="vertical" size="middle" style={{ width: '100%' }}>
          <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {t('nodes.tokenHint', { name: tokenResult.name })}
          </Typography.Paragraph>
          <div>
            <Typography.Text type="secondary">{t('nodes.deployServerUrlLabel')}: </Typography.Text>
            <Typography.Text code copyable>
              {serverUrl}
            </Typography.Text>
          </div>
          <Alert type="info" showIcon message={t('nodes.deployServerHint')} />

          {tokenResult.expiresAt && (
            <Typography.Text type="secondary">
              {t('nodes.expiresAt')}
              {tokenResult.expiresAt}
            </Typography.Text>
          )}

          <div>
            <Typography.Text strong>{t('nodes.copyToken')}</Typography.Text>
            <Input.TextArea
              readOnly
              value={tokenResult.token}
              autoSize={{ minRows: 3, maxRows: 6 }}
              style={{ fontFamily: 'monospace', marginTop: 8 }}
            />
            <Button
              style={{ marginTop: 8 }}
              icon={<CopyOutlined />}
              onClick={() => copyText(tokenResult.token, 'nodes.copied')}
            >
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
      ) : (
        <Form form={form} layout="vertical" initialValues={{ expires_in_hours: 168 }}>
          <Form.Item
            label={t('nodes.fieldName')}
            name="name"
            rules={[{ required: true, message: t('nodes.nameRequired') }]}
          >
            <Input placeholder={t('nodes.namePlaceholder')} maxLength={128} showCount />
          </Form.Item>
          <Form.Item
            label={t('nodes.fieldExpiresHours')}
            name="expires_in_hours"
            tooltip={t('nodes.expiresTooltip')}
            rules={[{ required: true, message: t('nodes.hoursRequired') }]}
          >
            <InputNumber min={0} max={8760} style={{ width: '100%' }} placeholder={t('nodes.expiresPlaceholder')} />
          </Form.Item>
        </Form>
      )}
    </Modal>
  );
}
