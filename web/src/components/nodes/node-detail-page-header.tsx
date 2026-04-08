'use client';

import { DeleteOutlined } from '@ant-design/icons';
import { App, Button, Card, Popconfirm, Space, Typography } from 'antd';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { NodeToolbar } from '@/components/nodes/node-toolbar';
import { useT } from '@/i18n';
import { deleteNode } from '@/services/node';

export function NodeDetailPageHeader({ id, name }: { id: number; name: string }) {
  const router = useRouter();
  const { message } = App.useApp();
  const t = useT();
  const [deleting, setDeleting] = useState(false);

  return (
    <Card className="page-card">
      <Space orientation="vertical" size={12} style={{ width: '100%' }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }} align="start" wrap>
          <div>
            <Typography.Title level={3} style={{ marginBottom: 0 }}>
              {t('nodes.detailTitle')}
            </Typography.Title>
            <Typography.Text type="secondary">{name}</Typography.Text>
          </div>
          <Popconfirm
            title={t('nodes.deleteNode')}
            description={t('nodes.deleteDetailConfirm', { name })}
            okText={t('common.delete')}
            cancelText={t('common.cancel')}
            okButtonProps={{ danger: true, loading: deleting }}
            onConfirm={async () => {
              setDeleting(true);
              try {
                await deleteNode(id);
                message.success(t('nodes.deleted', { name }));
                router.push('/nodes');
                router.refresh();
              } catch (e) {
                message.error(e instanceof Error ? e.message : t('nodes.deleteFailed'));
              } finally {
                setDeleting(false);
              }
            }}
          >
            <Button danger icon={<DeleteOutlined />} loading={deleting}>
              {t('nodes.deleteNode')}
            </Button>
          </Popconfirm>
        </Space>
        <NodeToolbar />
      </Space>
    </Card>
  );
}
