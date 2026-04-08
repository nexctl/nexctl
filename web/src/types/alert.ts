export interface AlertRule {
  id: number;
  name: string;
  type: string;
  target: string;
  enabled: boolean;
}

export interface AlertEvent {
  id: number;
  severity: string;
  node_name: string;
  summary: string;
  status: string;
  created_at: string;
}

