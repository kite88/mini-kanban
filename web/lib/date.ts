import { DEFAULT_LANG, LANG_LOCALES, type Lang } from "@/lib/i18n";

export function toInputValue(iso?: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(
    d.getMinutes(),
  )}`;
}

export function fromInputValue(v: string): string | null {
  if (!v) return null;
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? null : d.toISOString();
}

/** 语言 → Intl locale（与 <html lang> 同一套取值） */
export function localeOf(lang: Lang): string {
  return LANG_LOCALES[lang];
}

/** 语言未就绪（非组件代码）时的兜底 locale */
export const DEFAULT_LOCALE = LANG_LOCALES[DEFAULT_LANG];

/**
 * 展示型日期时间统一走这里，按当前界面语言格式化（如 2026/09/20 14:30 或
 * 09/20/2026, 14:30）。组件里请传 useI18n().locale。
 */
export function formatDateTime(iso?: string | null, locale: string = DEFAULT_LOCALE): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleString(locale, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

export function isOverdue(iso?: string | null): boolean {
  if (!iso) return false;
  return new Date(iso).getTime() < Date.now();
}
