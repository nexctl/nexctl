import type { AlertEvent, AlertRule } from '@/types/alert';
import type { AuditLogItem } from '@/types/audit';
import type { FileItem } from '@/types/file';
import type { NodeDetail, NodeItem } from '@/types/node';
import type { TaskItem } from '@/types/task';
import type { ReleaseItem } from '@/types/upgrade';

const runtime = {
  cpu_percent: 28.4,
  memory_percent: 61.2,
  disk_percent: 47.6,
  network_rx_bps: 103244,
  network_tx_bps: 60231,
  load_1: 0.68,
  load_5: 0.75,
  load_15: 0.71,
  uptime_seconds: 286602,
  process_count: 214,
  updated_at: '2026-04-08T10:30:00Z',
};

export const mockNodes: NodeItem[] = [
  {
    id: 1,
    name: 'prod-api-01',
    status: 'online',
    hostname: 'prod-api-01.internal',
    platform: 'Ubuntu 22.04',
    arch: 'x86_64',
    agent_version: '0.1.0',
    last_heartbeat_at: '2026-04-08T10:30:00Z',
    labels: ['prod', 'api', 'cn-sh'],
    runtime_state: runtime,
  },
  {
    id: 2,
    name: 'prod-worker-02',
    status: 'unstable',
    hostname: 'prod-worker-02.internal',
    platform: 'Debian 12',
    arch: 'x86_64',
    agent_version: '0.1.0',
    last_heartbeat_at: '2026-04-08T10:28:17Z',
    labels: ['prod', 'worker'],
    runtime_state: { ...runtime, cpu_percent: 72.3, memory_percent: 79.1 },
  },
  {
    id: 3,
    name: 'edge-macos-01',
    status: 'offline',
    hostname: 'edge-macos-01.local',
    platform: 'macOS 14',
    arch: 'arm64',
    agent_version: '0.1.0',
    last_heartbeat_at: '2026-04-08T09:10:12Z',
    labels: ['edge', 'darwin'],
    runtime_state: { ...runtime, cpu_percent: 0, memory_percent: 0, disk_percent: 0 },
  },
];

export const mockNodeDetail: NodeDetail = {
  ...mockNodes[0],
  services: [
    { name: 'nginx', status: 'running', startup_type: 'enabled' },
    { name: 'docker', status: 'running', startup_type: 'enabled' },
    { name: 'ssh', status: 'running', startup_type: 'enabled' },
  ],
  recent_tasks: [
    { id: 1001, type: 'shell_command', status: 'success', target: 'prod-api-01', created_at: '2026-04-08 10:20:00' },
    { id: 1002, type: 'upgrade_agent', status: 'running', target: 'prod-api-01', created_at: '2026-04-08 10:25:00' },
  ],
  alerts: [
    { id: 2001, severity: 'warning', summary: '磁盘使用率超过 80%', created_at: '2026-04-08 07:30:00' },
  ],
  short_term_metrics: [
    { time: '10:00', cpu: 22, memory: 60, disk: 47 },
    { time: '10:05', cpu: 31, memory: 62, disk: 47 },
    { time: '10:10', cpu: 27, memory: 63, disk: 47 },
    { time: '10:15', cpu: 36, memory: 61, disk: 47 },
    { time: '10:20', cpu: 29, memory: 62, disk: 47 },
    { time: '10:25', cpu: 28, memory: 61, disk: 47 },
  ],
};

export const mockTasks: TaskItem[] = [
  { id: 1001, type: 'shell_command', scope: '单节点', status: 'success', progress: 100, operator: 'admin', created_at: '2026-04-08 10:00:00', finished_at: '2026-04-08 10:00:05', detail: 'systemctl restart nginx' },
  { id: 1002, type: 'upload_file', scope: '标签:prod', status: 'running', progress: 58, operator: 'admin', created_at: '2026-04-08 10:10:00', detail: '/opt/releases/app.tar.gz' },
  { id: 1003, type: 'diagnostics', scope: '节点组:worker', status: 'pending', progress: 0, operator: 'ops.lead', created_at: '2026-04-08 10:26:00', detail: 'collect runtime and logs' },
];

export const mockFiles: FileItem[] = [
  { id: 1, file_name: 'agent-linux-amd64.tar.gz', size_text: '18.4 MB', checksum: '2e3f...c31a', created_at: '2026-04-08 08:00:00', distribution_count: 12 },
  { id: 2, file_name: 'nginx.conf', size_text: '12 KB', checksum: '08ad...f012', created_at: '2026-04-07 15:22:00', distribution_count: 4 },
];

export const mockReleases: ReleaseItem[] = [
  { id: 1, version: 'v0.1.0', channel: 'stable', notes: 'First internal release', created_at: '2026-04-08 09:00:00', rollout_status: 'available' },
  { id: 2, version: 'v0.1.1-rc1', channel: 'canary', notes: 'Fix reconnect and supervisor restart', created_at: '2026-04-08 10:00:00', rollout_status: 'rolling' },
];

export const mockAlertRules: AlertRule[] = [
  { id: 1, name: '节点离线', type: 'node_offline', target: 'all nodes', enabled: true },
  { id: 2, name: 'CPU 高负载', type: 'cpu_threshold', target: 'cpu > 85%', enabled: true },
];

export const mockAlertEvents: AlertEvent[] = [
  { id: 1, severity: 'critical', node_name: 'edge-macos-01', summary: '节点离线超过 5 分钟', status: 'open', created_at: '2026-04-08 09:15:00' },
  { id: 2, severity: 'warning', node_name: 'prod-worker-02', summary: 'CPU 使用率超过 85%', status: 'acknowledged', created_at: '2026-04-08 10:16:00' },
];

export const mockAuditLogs: AuditLogItem[] = [
  { id: 1, actor: 'admin', action: 'auth.login', resource: 'session/admin', detail: '控制台登录成功', created_at: '2026-04-08 09:00:00' },
  { id: 2, actor: 'admin', action: 'task.create', resource: 'task/1002', detail: '创建文件分发任务', created_at: '2026-04-08 10:10:00' },
  { id: 3, actor: 'ops.lead', action: 'node.upgrade', resource: 'node/1', detail: '触发 agent 灰度升级', created_at: '2026-04-08 10:25:00' },
];
