import { apiRequest } from '@/services/api';
import type { TaskItem } from '@/types/task';
import { mockTasks } from '@/utils/mock';

export async function getTasks() {
  const response = await apiRequest<TaskItem[] | { items: TaskItem[] }>('/tasks', undefined, { items: mockTasks });
  return Array.isArray(response) ? response : response.items;
}

/** 手动创建，或仅传 schedule_id 按计划任务模板立即执行一次 */
export type CreateTaskBody =
  | { schedule_id: number }
  | {
      task_type: string;
      scope_type: string;
      scope_value: string;
      detail: string;
    };

export type TaskScheduleItem = {
  id: number;
  name: string;
  cron_expr: string;
  task_type: string;
  scope: string;
  detail: string;
  enabled: boolean;
  next_run_at: string;
};

export async function getTaskSchedules() {
  const data = await apiRequest<{ items: TaskScheduleItem[] }>('/task-schedules', undefined, { items: [] });
  return data.items ?? [];
}

export type CreateScheduleBody = {
  name: string;
  cron_expr: string;
  task_type: string;
  scope_type: string;
  scope_value: string;
  detail: string;
  enabled?: boolean;
};

export async function createTaskSchedule(body: CreateScheduleBody) {
  return apiRequest<TaskScheduleItem>('/task-schedules', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export type TaskDetailResponse = {
  id: number;
  schedule_id?: number;
  type: string;
  scope: string;
  status: string;
  progress: number;
  operator: string;
  created_at: string;
  finished_at?: string;
  detail: string;
  output?: string;
  scope_type: string;
  scope_value: string;
};

export async function createTask(body: CreateTaskBody) {
  return apiRequest<TaskDetailResponse>('/tasks', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}
