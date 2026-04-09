'use client';

import { App, Form, Input, Modal, Select, Spin, Switch } from 'antd';
import { useCallback, useEffect, useState } from 'react';
import { useT } from '@/i18n';
import { createTaskSchedule } from '@/services/task';
import { getNodes } from '@/services/node';
import type { NodeItem } from '@/types/node';

type Props = {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
};

export function CreateScheduleModal({ open, onClose, onCreated }: Props) {
  const t = useT();
  const { message } = App.useApp();
  const [form] = Form.useForm<{
    name: string;
    cron_expr: string;
    task_type: string;
    scope_value: number[];
    detail: string;
    enabled: boolean;
  }>();
  const [nodes, setNodes] = useState<NodeItem[]>([]);
  const [loadingNodes, setLoadingNodes] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const taskType = Form.useWatch('task_type', form);

  const loadNodes = useCallback(async () => {
    setLoadingNodes(true);
    try {
      setNodes(await getNodes());
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('taskSchedules.loadNodesFailed'));
      setNodes([]);
    } finally {
      setLoadingNodes(false);
    }
  }, [message, t]);

  useEffect(() => {
    if (!open) {
      return;
    }
    form.resetFields();
    form.setFieldsValue({
      name: '',
      cron_expr: '0 * * * *',
      task_type: 'echo',
      scope_value: [],
      detail: '',
      enabled: true,
    });
    void loadNodes();
  }, [open, form, loadNodes]);

  const handleOk = async () => {
    try {
      const v = await form.validateFields();
      setSubmitting(true);
      const ids = [...(v.scope_value ?? [])].sort((a, b) => a - b);
      await createTaskSchedule({
        name: v.name?.trim() ?? '',
        cron_expr: v.cron_expr.trim(),
        task_type: v.task_type,
        scope_type: 'node',
        scope_value: ids.map(String).join(','),
        detail: v.detail?.trim() ?? '',
        enabled: v.enabled,
      });
      message.success(t('taskSchedules.createSuccess'));
      onCreated();
      onClose();
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) {
        return;
      }
      message.error(e instanceof Error ? e.message : t('taskSchedules.createFailed'));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title={t('taskSchedules.createModalTitle')}
      open={open}
      onCancel={onClose}
      onOk={handleOk}
      confirmLoading={submitting}
      destroyOnHidden
      width={560}
    >
      <Spin spinning={loadingNodes}>
        <Form form={form} layout="vertical" style={{ marginTop: 8 }}>
          <Form.Item name="name" label={t('taskSchedules.colName')}>
            <Input placeholder={t('taskSchedules.namePlaceholder')} />
          </Form.Item>
          <Form.Item
            name="cron_expr"
            label={t('taskSchedules.colCron')}
            extra={t('taskSchedules.cronHint')}
            rules={[{ required: true, message: t('taskSchedules.cronRequired') }]}
          >
            <Input placeholder="0 * * * *" />
          </Form.Item>
          <Form.Item
            name="task_type"
            label={t('tasks.newTaskType')}
            rules={[{ required: true, message: t('tasks.newTaskTypeRequired') }]}
          >
            <Select
              options={[
                { value: 'echo', label: t('tasks.newTaskTypeEcho') },
                { value: 'shell_command', label: t('tasks.newTaskTypeShell') },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="scope_value"
            label={t('taskSchedules.nodesLabel')}
            rules={[
              { required: true, message: t('taskSchedules.nodesRequired') },
              { type: 'array', min: 1, message: t('taskSchedules.nodesRequired') },
            ]}
          >
            <Select
              mode="multiple"
              allowClear
              showSearch
              optionFilterProp="label"
              placeholder={t('taskSchedules.nodesPlaceholder')}
              maxTagCount="responsive"
              options={nodes.map((n) => ({
                value: n.id,
                label: `${n.name} (#${n.id})`,
              }))}
            />
          </Form.Item>
          <Form.Item
            name="detail"
            label={t('tasks.newTaskDetail')}
            rules={
              taskType === 'shell_command'
                ? [{ required: true, message: t('tasks.newTaskDetailRequiredShell') }]
                : []
            }
          >
            <Input.TextArea
              rows={3}
              placeholder={
                taskType === 'shell_command'
                  ? t('tasks.newTaskDetailPlaceholderShell')
                  : t('tasks.newTaskDetailPlaceholderEcho')
              }
            />
          </Form.Item>
          <Form.Item name="enabled" label={t('taskSchedules.colEnabled')} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Spin>
    </Modal>
  );
}
