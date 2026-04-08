import type { AppLocale } from '@/i18n/types';

import en from '@/i18n/messages/en.json';
import zhCN from '@/i18n/messages/zh-CN.json';

const catalog: Record<AppLocale, Record<string, unknown>> = {
  'zh-CN': zhCN as Record<string, unknown>,
  en: en as Record<string, unknown>,
};

export function loadMessages(locale: AppLocale): Record<string, unknown> {
  return catalog[locale] ?? catalog['zh-CN'];
}

export function getNested(obj: Record<string, unknown>, path: string): string | undefined {
  const parts = path.split('.');
  let cur: unknown = obj;
  for (const p of parts) {
    if (cur && typeof cur === 'object' && p in (cur as Record<string, unknown>)) {
      cur = (cur as Record<string, unknown>)[p];
    } else {
      return undefined;
    }
  }
  return typeof cur === 'string' ? cur : undefined;
}

export function interpolate(template: string, params?: Record<string, string | number>): string {
  if (!params) return template;
  return template.replace(/\{\{(\w+)\}\}/g, (_, key: string) => String(params[key] ?? ''));
}
