'use client';

import { Card, Col, Descriptions, Progress, Row, Space, Table, Tag, Typography } from 'antd';
import { useMemo } from 'react';
import { MiniLineChart } from '@/components/charts/mini-line-chart';
import { useT } from '@/i18n';
import type { NodeDetail } from '@/types/node';

function formatStatus(value: string, t: ReturnType<typeof useT>) {
  const k = `nodes.status.${value}`;
  const s = t(k);
  return s === k ? value : s;
}

export function NodeSummaryCards({ detail }: { detail: NodeDetail }) {
  const t = useT();

  const serviceColumns = useMemo(
    () => [
      { title: t('nodes.colServiceName'), dataIndex: 'name', key: 'name' },
      { title: t('nodes.labelStatus'), dataIndex: 'status', key: 'status' },
      { title: t('nodes.colStartup'), dataIndex: 'startup_type', key: 'startup_type' },
    ],
    [t],
  );

  const taskColumns = useMemo(
    () => [
      { title: t('nodes.colTask'), dataIndex: 'type', key: 'type' },
      { title: t('nodes.labelStatus'), dataIndex: 'status', key: 'status' },
      { title: t('nodes.colTime'), dataIndex: 'created_at', key: 'created_at' },
    ],
    [t],
  );

  const alertColumns = useMemo(
    () => [
      { title: t('nodes.colLevel'), dataIndex: 'severity', key: 'severity' },
      { title: t('nodes.colSummary'), dataIndex: 'summary', key: 'summary' },
      { title: t('nodes.colTime'), dataIndex: 'created_at', key: 'created_at' },
    ],
    [t],
  );

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      <Row gutter={16}>
        <Col xs={24} xl={12}>
          <Card className="page-card" title={t('nodes.summaryBasic')}>
            <Descriptions column={2} size="small">
              <Descriptions.Item label={t('nodes.labelName')}>{detail.name}</Descriptions.Item>
              <Descriptions.Item label={t('nodes.labelStatus')}>
                <Tag color={detail.status === 'online' ? 'green' : detail.status === 'unstable' ? 'orange' : 'red'}>
                  {formatStatus(detail.status, t)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Hostname">{detail.hostname}</Descriptions.Item>
              <Descriptions.Item label={t('nodes.labelPlatform')}>{detail.platform}</Descriptions.Item>
              <Descriptions.Item label={t('nodes.labelArch')}>{detail.arch}</Descriptions.Item>
              <Descriptions.Item label={t('nodes.labelAgentVer')}>{detail.agent_version}</Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
        <Col xs={24} xl={12}>
          <Card className="page-card" title={t('nodes.summaryState')}>
            <Space orientation="vertical" size={12} style={{ width: '100%' }}>
              <div>
                <Typography.Text>CPU</Typography.Text>
                <Progress percent={Math.round(detail.runtime_state?.cpu_percent ?? 0)} />
              </div>
              <div>
                <Typography.Text>{t('dashboard.colMemory')}</Typography.Text>
                <Progress
                  percent={Math.round(detail.runtime_state?.memory_percent ?? 0)}
                  status="active"
                />
              </div>
              <div>
                <Typography.Text>{t('nodes.tableDisk')}</Typography.Text>
                <Progress
                  percent={Math.round(detail.runtime_state?.disk_percent ?? 0)}
                  strokeColor="#d97706"
                />
              </div>
            </Space>
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} xl={12}>
          <Card className="page-card" title={t('nodes.summaryCpuMemTrend')}>
            <MiniLineChart
              values={(detail.short_term_metrics ?? []).map((item) => item.cpu)}
              color="#1677ff"
            />
            <MiniLineChart
              values={(detail.short_term_metrics ?? []).map((item) => item.memory)}
              color="#16a34a"
            />
          </Card>
        </Col>
        <Col xs={24} xl={12}>
          <Card className="page-card" title={t('nodes.summaryServices')}>
            <Table rowKey="name" pagination={false} size="small" dataSource={detail.services} columns={serviceColumns} />
          </Card>
        </Col>
      </Row>

      <div className="section-grid">
        <Card className="page-card" title={t('nodes.summaryTasks')}>
          <Table
            rowKey="id"
            size="small"
            pagination={false}
            dataSource={detail.recent_tasks}
            columns={taskColumns}
          />
        </Card>
        <Card className="page-card" title={t('nodes.summaryAlerts')}>
          <Table rowKey="id" size="small" pagination={false} dataSource={detail.alerts} columns={alertColumns} />
        </Card>
      </div>
    </Space>
  );
}
