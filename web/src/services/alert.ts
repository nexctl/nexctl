import { apiRequest } from '@/services/api';
import type { AlertEvent, AlertRule } from '@/types/alert';
import { mockAlertEvents, mockAlertRules } from '@/utils/mock';

export async function getAlertRules() {
  const response = await apiRequest<AlertRule[] | { items: AlertRule[] }>('/alerts/rules', undefined, { items: mockAlertRules });
  return Array.isArray(response) ? response : response.items;
}

export async function getAlertEvents() {
  const response = await apiRequest<AlertEvent[] | { items: AlertEvent[] }>('/alerts/events', undefined, { items: mockAlertEvents });
  return Array.isArray(response) ? response : response.items;
}
