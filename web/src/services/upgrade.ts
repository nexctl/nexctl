import { apiRequest } from '@/services/api';
import type { ReleaseItem } from '@/types/upgrade';
import { mockReleases } from '@/utils/mock';

export async function getReleases() {
  const response = await apiRequest<ReleaseItem[] | { items: ReleaseItem[] }>('/upgrades/releases', undefined, { items: mockReleases });
  return Array.isArray(response) ? response : response.items;
}
