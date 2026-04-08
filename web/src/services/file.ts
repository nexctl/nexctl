import { apiRequest } from '@/services/api';
import type { FileItem } from '@/types/file';
import { mockFiles } from '@/utils/mock';

export async function getFiles() {
  const response = await apiRequest<FileItem[] | { items: FileItem[] }>('/files', undefined, { items: mockFiles });
  return Array.isArray(response) ? response : response.items;
}
