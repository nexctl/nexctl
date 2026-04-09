export type TaskStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled';

export interface TaskItem {
  id: number;
  /** 非空表示由计划任务触发的执行实例 */
  schedule_id?: number;
  type: string;
  scope: string;
  status: TaskStatus;
  progress: number;
  operator: string;
  created_at: string;
  finished_at?: string;
  detail: string;
}

