'use client';

import { App, Button, Form, Input, Modal, Space } from 'antd';
import { useState } from 'react';
import { useT } from '@/i18n';
import { createPendingNode } from '@/services/node';
import { EnrollmentDeployCommands } from '@/components/nodes/enrollment-deploy-commands';

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
  const [credResult, setCredResult] = useState<{
    agentId: string;
    agentSecret: string;
    nodeKey: string;
    name: string;
    nodeId: number;
  } | null>(null);

  const handleClose = () => {
    setCredResult(null);
    form.resetFields();
    onClose();
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    setSubmitting(true);
    try {
      const res = await createPendingNode({
        name: values.name.trim(),
      });
      setCredResult({
        agentId: res.agent_id,
        agentSecret: res.agent_secret,
        nodeKey: res.node_key,
        name: res.name,
        nodeId: res.id,
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
      title={credResult ? t('nodes.modalCredentialTitle') : t('nodes.modalAddTitle')}
      open={open}
      onCancel={handleClose}
      footer={
        credResult ? (
          <Button type="primary" onClick={handleClose}>
            {t('common.close')}
          </Button>
        ) : (
          <Space>
            <Button onClick={handleClose}>{t('common.cancel')}</Button>
            <Button type="primary" loading={submitting} onClick={handleSubmit}>
              {t('nodes.createNode')}
            </Button>
          </Space>
        )
      }
      destroyOnHidden
      width={720}
    >
      {credResult ? (
        <EnrollmentDeployCommands
          nodeName={credResult.name}
          agentId={credResult.agentId}
          agentSecret={credResult.agentSecret}
          nodeKey={credResult.nodeKey}
          nodeId={credResult.nodeId}
        />
      ) : (
        <Form form={form} layout="vertical">
          <Form.Item
            label={t('nodes.fieldName')}
            name="name"
            rules={[{ required: true, message: t('nodes.nameRequired') }]}
          >
            <Input placeholder={t('nodes.namePlaceholder')} maxLength={128} showCount />
          </Form.Item>
        </Form>
      )}
    </Modal>
  );
}
