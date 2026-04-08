'use client';

import { PlusOutlined } from '@ant-design/icons';
import { Button, Space } from 'antd';
import { useState } from 'react';
import { AddNodeModal } from '@/components/nodes/add-node-modal';
import { useT } from '@/i18n';
import { useRouter } from 'next/navigation';

type Props = {
  onNodeCreated?: () => void;
};

export function NodesToolbar({ onNodeCreated }: Props) {
  const router = useRouter();
  const t = useT();
  const [open, setOpen] = useState(false);

  return (
    <>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>
          {t('nodes.addNode')}
        </Button>
      </Space>
      <AddNodeModal
        open={open}
        onClose={() => setOpen(false)}
        onCreated={() => {
          onNodeCreated?.();
          router.refresh();
        }}
      />
    </>
  );
}
