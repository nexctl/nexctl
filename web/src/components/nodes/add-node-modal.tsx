'use client';

import { CopyOutlined } from '@ant-design/icons';
import { App, Button, Form, Input, InputNumber, Modal, Space, Typography } from 'antd';
import { useState } from 'react';
import { useT } from '@/i18n';
import { createPendingNode } from '@/services/node';

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
      width={560}
    >
      {tokenResult ? (
        <Space orientation="vertical" size="middle" style={{ width: '100%' }}>
          <Typography.Text type="secondary">
            {t('nodes.tokenHint', { name: tokenResult.name })}{' '}
            <Typography.Text code>OPSPILOT_AGENT_ENROLLMENT_TOKEN</Typography.Text> {t('nodes.tokenHintOr')}{' '}
            <Typography.Text code>enrollment_token</Typography.Text>。
          </Typography.Text>
          {tokenResult.expiresAt && (
            <Typography.Text type="secondary">
              {t('nodes.expiresAt')}
              {tokenResult.expiresAt}
            </Typography.Text>
          )}
          <Input.TextArea
            readOnly
            value={tokenResult.token}
            autoSize={{ minRows: 3, maxRows: 6 }}
            style={{ fontFamily: 'monospace' }}
          />
          <Button
            icon={<CopyOutlined />}
            onClick={async () => {
              try {
                await navigator.clipboard.writeText(tokenResult.token);
                message.success(t('nodes.copied'));
              } catch {
                message.warning(t('nodes.copyFailed'));
              }
            }}
          >
            {t('nodes.copyToken')}
          </Button>
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
