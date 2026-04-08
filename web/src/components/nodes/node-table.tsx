'use client';

import {
  CloudDownloadOutlined,
  CodeOutlined,
  DeleteOutlined,
  FileSyncOutlined,
  SendOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { App, Button, Popconfirm, Space, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useRouter } from 'next/navigation';
import { useMemo, useState } from 'react';
import { useT } from '@/i18n';
import { NodeInstallModal } from '@/components/nodes/node-install-modal';
import { NodeTerminalModal } from '@/components/nodes/node-terminal-modal';
import { deleteNode } from '@/services/node';
import type { NodeItem } from '@/types/node';
import { formatDateTimeLocal } from '@/utils/datetime';

function formatStatus(value: string, t: ReturnType<typeof useT>) {
  const k = `nodes.status.${value}`;
  const s = t(k);
  return s === k ? value : s;
}

type NodeTableProps = {
  nodes: NodeItem[];
  /** 删除成功后刷新列表（客户端拉数据时使用，替代 router.refresh） */
  onAfterDelete?: () => void;
};

export function NodeTable({ nodes, onAfterDelete }: NodeTableProps) {
  const router = useRouter();
  const { message } = App.useApp();
  const t = useT();
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [installNode, setInstallNode] = useState<NodeItem | null>(null);
  const [terminalNode, setTerminalNode] = useState<NodeItem | null>(null);

  const columns: ColumnsType<NodeItem> = useMemo(
    () => [
      { title: t('nodes.tableName'), dataIndex: 'name', key: 'name', fixed: 'left', width: 160 },
      {
        title: t('nodes.tableStatus'),
        dataIndex: 'status',
        key: 'status',
        width: 110,
        render: (value: NodeItem['status']) => (
          <Tag
            color={
              value === 'online'
                ? 'green'
                : value === 'pending'
                  ? 'blue'
                  : value === 'unstable'
                    ? 'orange'
                    : 'red'
            }
          >
            {formatStatus(value, t)}
          </Tag>
        ),
      },
      { title: 'Hostname', dataIndex: 'hostname', key: 'hostname', width: 200 },
      {
        title: t('nodes.tableOsArch'),
        key: 'platform',
        width: 180,
        render: (_, record) => `${record.platform} / ${record.arch}`,
      },
      { title: t('nodes.tableAgentVer'), dataIndex: 'agent_version', key: 'agent_version', width: 110 },
      {
        title: 'CPU',
        key: 'cpu',
        width: 90,
        render: (_, row) => `${(row.runtime_state?.cpu_percent ?? 0).toFixed(1)}%`,
      },
      {
        title: t('nodes.tableMem'),
        key: 'memory',
        width: 90,
        render: (_, row) => `${(row.runtime_state?.memory_percent ?? 0).toFixed(1)}%`,
      },
      {
        title: t('nodes.tableDisk'),
        key: 'disk',
        width: 90,
        render: (_, row) => `${(row.runtime_state?.disk_percent ?? 0).toFixed(1)}%`,
      },
      {
        title: t('nodes.tableHeartbeat'),
        dataIndex: 'last_heartbeat_at',
        key: 'last_heartbeat_at',
        width: 200,
        render: (value: string) => formatDateTimeLocal(value),
      },
      {
        title: t('nodes.tableTags'),
        key: 'labels',
        width: 220,
        render: (_, row) => (
          <Space wrap>
            {row.labels.map((label) => (
              <Tag key={label}>{label}</Tag>
            ))}
          </Space>
        ),
      },
      {
        title: t('nodes.tableActions'),
        key: 'actions',
        fixed: 'right',
        width: 340,
        render: (_, row) => (
          <Space wrap>
            <Button size="small" icon={<CloudDownloadOutlined />} onClick={() => setInstallNode(row)}>
              {t('nodes.install')}
            </Button>
            <Button
              size="small"
              icon={<CodeOutlined />}
              onClick={() => message.info(t('nodes.featureComingSoon'))}
            >
              {t('nodes.cmd')}
            </Button>
            <Button
              size="small"
              icon={<SendOutlined />}
              onClick={() => setTerminalNode(row)}
            >
              {t('nodes.terminal')}
            </Button>
            <Button
              size="small"
              icon={<FileSyncOutlined />}
              onClick={() => message.info(t('nodes.featureComingSoon'))}
            >
              {t('nodes.file')}
            </Button>
            <Button
              size="small"
              icon={<ThunderboltOutlined />}
              onClick={() => router.push(`/nodes/${row.id}`)}
            >
              {t('nodes.detail')}
            </Button>
            <Popconfirm
              title={t('nodes.deleteNode')}
              description={t('nodes.deleteConfirm', { name: row.name })}
              okText={t('common.delete')}
              cancelText={t('common.cancel')}
              okButtonProps={{ danger: true, loading: deletingId === row.id }}
              onConfirm={async () => {
                setDeletingId(row.id);
                try {
                  await deleteNode(row.id);
                  message.success(t('nodes.deleted', { name: row.name }));
                  onAfterDelete?.();
                  router.refresh();
                } catch (e) {
                  message.error(e instanceof Error ? e.message : t('nodes.deleteFailed'));
                } finally {
                  setDeletingId(null);
                }
              }}
            >
              <Button size="small" danger icon={<DeleteOutlined />} disabled={deletingId !== null}>
                {t('common.delete')}
              </Button>
            </Popconfirm>
          </Space>
        ),
      },
    ],
    [t, deletingId, message, router, onAfterDelete],
  );

  return (
    <>
      <Table rowKey="id" columns={columns} dataSource={nodes} scroll={{ x: 1500 }} />
      <NodeInstallModal
        open={installNode !== null}
        onClose={() => setInstallNode(null)}
        node={installNode}
      />
      {terminalNode !== null ? (
        <NodeTerminalModal
          open
          onClose={() => setTerminalNode(null)}
          nodeId={terminalNode.id}
          nodeName={terminalNode.name}
        />
      ) : null}
    </>
  );
}
