'use client';

import { App, Form, Input, Modal, Segmented, Select, Spin } from 'antd';
import { useCallback, useEffect, useState } from 'react';
import { useT } from '@/i18n';
import { createTask, getTaskSchedules, type TaskScheduleItem } from '@/services/task';
import { getNodes } from '@/services/node';
import type { NodeItem } from '@/types/node';

type TriggerMode = 'manual' | 'schedule';

type FormValues = {
  trigger: TriggerMode;
  schedule_id?: number;
  task_type: string;
  scope_value: string;
  detail: string;
};

type Props = {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
};

export function NewTaskModal({ open, onClose, onCreated }: Props) {
  const t = useT();
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [nodes, setNodes] = useState<NodeItem[]>([]);
  const [schedules, setSchedules] = useState<TaskScheduleItem[]>([]);
  const [loadingNodes, setLoadingNodes] = useState(false);
  const [loadingSchedules, setLoadingSchedules] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const trigger = Form.useWatch('trigger', form);
  const taskType = Form.useWatch('task_type', form);

  const loadNodes = useCallback(async () => {
    setLoadingNodes(true);
    try {
      const list = await getNodes();
      setNodes(list);
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('tasks.newTaskLoadNodesFailed'));
      setNodes([]);
    } finally {
      setLoadingNodes(false);
    }
  }, [message, t]);

  const loadSchedules = useCallback(async () => {
    setLoadingSchedules(true);
    try {
      const list = await getTaskSchedules();
      setSchedules(list);
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('tasks.newTaskLoadSchedulesFailed'));
      setSchedules([]);
    } finally {
      setLoadingSchedules(false);
    }
  }, [message, t]);

  useEffect(() => {
    if (!open) {
      return;
    }
    form.resetFields();
    form.setFieldsValue({
      trigger: 'manual',
      task_type: 'echo',
      scope_value: undefined,
      detail: '',
      schedule_id: undefined,
    });
    void loadNodes();
    void loadSchedules();
  }, [open, form, loadNodes, loadSchedules]);

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      if (values.trigger === 'schedule') {
        await createTask({ schedule_id: values.schedule_id! });
      } else {
        await createTask({
          task_type: values.task_type,
          scope_type: 'node',
          scope_value: String(values.scope_value),
          detail: values.detail?.trim() ?? '',
        });
      }
      message.success(t('tasks.newTaskSuccess'));
      onCreated();
      onClose();
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) {
        return;
      }
      message.error(e instanceof Error ? e.message : t('tasks.newTaskFailed'));
    } finally {
      setSubmitting(false);
    }
  };

  const loading = loadingNodes || loadingSchedules;

  return (
    <Modal
      title={t('tasks.newTaskModalTitle')}
      open={open}
      onCancel={onClose}
      onOk={handleOk}
      confirmLoading={submitting}
      destroyOnHidden
      width={560}
    >
      <Spin spinning={loading}>
        <Form form={form} layout="vertical" style={{ marginTop: 8 }}>
          <Form.Item name="trigger" label={t('tasks.newTaskTrigger')}>
            <Segmented
              block
              options={[
                { value: 'manual', label: t('tasks.newTaskTriggerManual') },
                { value: 'schedule', label: t('tasks.newTaskTriggerSchedule') },
              ]}
            />
          </Form.Item>

          {trigger === 'schedule' ? (
            <Form.Item
              name="schedule_id"
              label={t('tasks.newTaskPickSchedule')}
              rules={[{ required: true, message: t('tasks.newTaskScheduleRequired') }]}
            >
              <Select
                showSearch
                optionFilterProp="label"
                placeholder={t('tasks.newTaskPickSchedulePlaceholder')}
                notFoundContent={schedules.length === 0 && !loading ? t('tasks.newTaskNoSchedules') : undefined}
                options={schedules.map((s) => ({
                  value: s.id,
                  label: `${s.name || t('tasks.newTaskScheduleUnnamed')} (#${s.id}) · ${s.cron_expr}`,
                  disabled: !s.enabled,
                }))}
              />
            </Form.Item>
          ) : (
            <>
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
                label={t('tasks.newTaskNode')}
                rules={[{ required: true, message: t('tasks.newTaskNodeRequired') }]}
              >
                <Select
                  showSearch
                  optionFilterProp="label"
                  placeholder={t('tasks.newTaskNodePlaceholder')}
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
                  rows={4}
                  placeholder={
                    taskType === 'shell_command'
                      ? t('tasks.newTaskDetailPlaceholderShell')
                      : t('tasks.newTaskDetailPlaceholderEcho')
                  }
                />
              </Form.Item>
            </>
          )}
        </Form>
      </Spin>
    </Modal>
  );
}
