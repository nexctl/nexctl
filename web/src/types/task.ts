export type TaskStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled';

export interface TaskItem {
  id: number;
  type: string;
  scope: string;
  status: TaskStatus;
  progress: number;
  operator: string;
  created_at: string;
  finished_at?: string;
  detail: string;
}

