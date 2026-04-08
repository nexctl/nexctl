import { apiRequest } from '@/services/api';
import type { NodeDetail, NodeItem } from '@/types/node';
import { mockNodeDetail, mockNodes } from '@/utils/mock';

export async function getNodes() {
  const response = await apiRequest<NodeItem[] | { items: NodeItem[] }>('/nodes', undefined, { items: mockNodes });
  return Array.isArray(response) ? response : response.items;
}

export function getNodeDetail(id: string) {
  return apiRequest<NodeDetail>(`/nodes/${id}`, undefined, { ...mockNodeDetail, id: Number(id) });
}

export async function deleteNode(id: number | string) {
  await apiRequest<{ deleted: boolean }>(`/nodes/${id}`, { method: 'DELETE' });
}

export type CreatePendingNodeResult = {
  id: number;
  name: string;
  status: string;
  enrollment_token: string;
  enrollment_expires_at?: string;
};

export async function createPendingNode(body: { name: string; expires_in_hours?: number }) {
  return apiRequest<CreatePendingNodeResult>('/nodes', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

