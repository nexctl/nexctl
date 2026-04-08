'use client';

import { useCallback, useEffect, useState } from 'react';
import { getNodes } from '@/services/node';
import type { NodeItem } from '@/types/node';

/** 在客户端拉取节点列表（依赖 localStorage 中的 Bearer token，不可在 RSC 中调用）。 */
export function useNodes() {
  const [nodes, setNodes] = useState<NodeItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const list = await getNodes();
      setNodes(list);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'load failed');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refetch();
  }, [refetch]);

  return { nodes, loading, error, refetch };
}
