"use client";

import { LANGS, LANG_LABELS, LANG_SHORT_LABELS } from "@/lib/i18n";
import { useI18n } from "@/components/I18nProvider";

/**
 * 语言切换：与主题切换同一套外观（分段式单选）。
 * 标签用「中文 / EN」而不是全称，导航栏里更省地方。
 * 不需要 mounted 守卫：语言来自服务端读到的 cookie，SSR 与首次渲染一致。
 */
export function LanguageToggle({ className = "" }: { className?: string }) {
  const { lang, setLang, t } = useI18n();

  return (
    <div
      role="radiogroup"
      aria-label={t("language.label")}
      title={t("language.label")}
      className={`inline-flex items-center gap-0.5 rounded-lg border border-slate-200 bg-white p-0.5 dark:border-slate-700 dark:bg-slate-900 ${className}`}
    >
      {LANGS.map((value) => {
        const active = lang === value;
        return (
          <button
            key={value}
            type="button"
            role="radio"
            aria-checked={active}
            title={LANG_LABELS[value]}
            aria-label={LANG_LABELS[value]}
            onClick={() => setLang(value)}
            className={`min-w-9 rounded-md px-2 py-1 text-xs leading-5 transition ${
              active
                ? "bg-indigo-600 text-white"
                : "text-slate-500 hover:bg-slate-100 hover:text-slate-700 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-200"
            }`}
          >
            {LANG_SHORT_LABELS[value]}
          </button>
        );
      })}
    </div>
  );
}
