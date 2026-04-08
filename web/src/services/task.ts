import { apiRequest } from '@/services/api';
import type { TaskItem } from '@/types/task';
import { mockTasks } from '@/utils/mock';

export async function getTasks() {
  const response = await apiRequest<TaskItem[] | { items: TaskItem[] }>('/tasks', undefined, { items: mockTasks });
  return Array.isArray(response) ? response : response.items;
}
