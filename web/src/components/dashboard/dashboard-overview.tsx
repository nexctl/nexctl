'use client';

import { Card, Col, Progress, Row, Statistic, Table, Tag } from 'antd';
import { useMemo } from 'react';
import { useT } from '@/i18n';
import type { NodeItem } from '@/types/node';

function statusLabel(status: string, t: ReturnType<typeof useT>) {
  const k = `nodes.status.${status}`;
  const s = t(k);
  return s === k ? status : s;
}

export function DashboardOverview({ nodes }: { nodes: NodeItem[] }) {
  const t = useT();
  const onlineCount = nodes.filter((item) => item.status === 'online').length;
  const unstableCount = nodes.filter((item) => item.status === 'unstable').length;
  const offlineCount = nodes.filter((item) => item.status === 'offline').length;

  const tableColumns = useMemo(
    () => [
      { title: t('dashboard.colNode'), dataIndex: 'name', key: 'name' },
      {
        title: t('dashboard.colCpu'),
        key: 'cpu',
        render: (_: unknown, row: NodeItem) => `${row.runtime_state.cpu_percent.toFixed(1)}%`,
      },
      {
        title: t('dashboard.colMemory'),
        key: 'memory',
        render: (_: unknown, row: NodeItem) => `${row.runtime_state.memory_percent.toFixed(1)}%`,
      },
    ],
    [t],
  );

  return (
    <>
      <div className="stat-grid" style={{ marginBottom: 16 }}>
        <Card className="page-card">
          <Statistic title={t('dashboard.totalNodes')} value={nodes.length} />
        </Card>
        <Card className="page-card">
          <Statistic title={t('dashboard.onlineNodes')} value={onlineCount} styles={{ content: { color: '#16a34a' } }} />
        </Card>
        <Card className="page-card">
          <Statistic title={t('dashboard.unstableNodes')} value={unstableCount} styles={{ content: { color: '#d97706' } }} />
        </Card>
        <Card className="page-card">
          <Statistic title={t('dashboard.offlineNodes')} value={offlineCount} styles={{ content: { color: '#dc2626' } }} />
        </Card>
      </div>
      <Row gutter={16}>
        <Col xs={24} xl={14}>
          <Card className="page-card" title={t('dashboard.resourceOverview')}>
            {nodes.slice(0, 3).map((node) => (
              <div key={node.id} style={{ marginBottom: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
                  <span>{node.name}</span>
                  <Tag color={node.status === 'online' ? 'green' : node.status === 'unstable' ? 'orange' : 'red'}>
                    {statusLabel(node.status, t)}
                  </Tag>
                </div>
                <Progress percent={Math.round(node.runtime_state.cpu_percent)} size="small" status="active" />
              </div>
            ))}
          </Card>
        </Col>
        <Col xs={24} xl={10}>
          <Card className="page-card" title={t('dashboard.recentNodeStatus')}>
            <Table rowKey="id" pagination={false} size="small" dataSource={nodes} columns={tableColumns} />
          </Card>
        </Col>
      </Row>
    </>
  );
}
