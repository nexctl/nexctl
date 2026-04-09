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

/** 控制台创建节点后返回的固定凭据（与数据库中一致）。 */
export type CreatePendingNodeResult = {
  id: number;
  name: string;
  status: string;
  agent_id: string;
  agent_secret: string;
  node_key: string;
  ws_url: string;
};

export async function createPendingNode(body: { name: string }) {
  return apiRequest<CreatePendingNodeResult>('/nodes', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

/** 查询节点固定接入凭据（需登录；用于「安装」弹窗）。 */
export type NodeAgentCredentials = {
  agent_id: string;
  agent_secret: string;
  node_key: string;
  ws_url: string;
};

export function getNodeAgentCredentials(id: number | string) {
  return apiRequest<NodeAgentCredentials>(`/nodes/${id}/agent-credentials`);
}

export type TriggerAgentUpgradeResult = {
  queued: boolean;
  request_id: string;
};

/** 向在线 Agent 下发升级检查指令（GitHub 有新版本则自更新并重启） */
export function triggerAgentUpgrade(id: number | string) {
  return apiRequest<TriggerAgentUpgradeResult>(`/nodes/${id}/upgrade-agent`, { method: 'POST' });
}

/** 远程文件操作（经 Agent WebSocket，需节点在线） */
export type NodeFileOpRequest = {
  op: string;
  path?: string;
  path_to?: string;
  content_b64?: string;
  max_bytes?: number;
  recursive?: boolean;
};

export type NodeFileEntry = {
  name: string;
  is_dir: boolean;
  size: number;
  mod_time: string;
};

export type NodeFileOpResult = {
  ok: boolean;
  error?: string;
  entries?: NodeFileEntry[];
  content_b64?: string;
  size?: number;
  mod_time?: string;
  is_dir?: boolean;
};

export async function nodeFileOp(id: number | string, body: NodeFileOpRequest) {
  return apiRequest<NodeFileOpResult>(`/nodes/${id}/file-op`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}
