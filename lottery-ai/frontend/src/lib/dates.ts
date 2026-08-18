/** 开奖日只取日历日期，避免 ISO 被切成 UTC 前一天。 */
export function formatDrawDate(value?: string | null): string {
  if (!value) return '-';
  const m = String(value).match(/^(\d{4}-\d{2}-\d{2})/);
  if (m) return m[1];
  return String(value);
}

export function formatYuan(n: number | undefined | null, signed = false): string {
  const v = Number(n || 0);
  const sign = signed && v > 0 ? '+' : '';
  return `${sign}${v.toFixed(0)}元`;
}
