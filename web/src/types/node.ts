export type NodeStatus = 'pending' | 'online' | 'unstable' | 'offline';

export interface RuntimeState {
  cpu_percent: number;
  memory_percent: number;
  disk_percent: number;
  network_rx_bps: number;
  network_tx_bps: number;
  load_1: number;
  load_5: number;
  load_15: number;
  uptime_seconds: number;
  process_count: number;
  updated_at: string;
}

export interface NodeItem {
  id: number;
  node_id?: number;
  name: string;
  status: NodeStatus;
  hostname: string;
  platform: string;
  arch: string;
  agent_version: string;
  last_heartbeat_at: string;
  labels: string[];
  runtime_state: RuntimeState;
}

export interface ServiceInfo {
  name: string;
  status: string;
  startup_type: string;
}

export interface TaskSummary {
  id: number;
  type: string;
  status: string;
  target: string;
  created_at: string;
}

export interface AlertSummary {
  id: number;
  severity: string;
  summary: string;
  created_at: string;
}

export interface NodeDetail extends NodeItem {
  services: ServiceInfo[];
  recent_tasks: TaskSummary[];
  alerts: AlertSummary[];
  short_term_metrics: Array<{ time: string; cpu: number; memory: number; disk: number }>;
}

