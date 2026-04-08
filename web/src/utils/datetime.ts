/**
 * 将 API 返回的 ISO 8601 时间格式化为浏览器本地时区、本地区域习惯（与浏览器语言一致）。
 */
export function formatDateTimeLocal(iso: string | undefined | null): string {
  if (iso == null || String(iso).trim() === '') {
    return '—';
  }
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) {
    return String(iso);
  }
  return d.toLocaleString(undefined, {
    dateStyle: 'short',
    timeStyle: 'medium',
  });
}
