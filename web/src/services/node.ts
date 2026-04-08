import { apiRequest } from '@/services/api';
import type { NodeDetail, NodeItem } from '@/types/node';

export async function getNodes() {
  const data = await apiRequest<{ items: NodeItem[] }>('/nodes');
  return data.items ?? [];
}

export function getNodeDetail(id: string) {
  return apiRequest<NodeDetail>(`/nodes/${id}`);
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

/** 为待注册节点重新签发注册令牌并返回（用于列表「安装」等场景；会作废此前未使用的旧令牌哈希） */
export function issueNodeEnrollmentToken(id: number | string) {
  return apiRequest<CreatePendingNodeResult>(`/nodes/${id}/enrollment-token`, { method: 'POST' });
}
