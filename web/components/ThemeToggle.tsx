"use client";

import { useEffect, useState } from "react";
import { useI18n } from "@/components/I18nProvider";
import { useTheme, type Theme } from "@/components/ThemeProvider";
import type { MessageKey } from "@/lib/i18n";

function SunIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" className="h-4 w-4">
      <circle cx="12" cy="12" r="4" />
      <path
        strokeLinecap="round"
        d="M12 2.5v2M12 19.5v2M4.6 4.6l1.4 1.4M18 18l1.4 1.4M2.5 12h2M19.5 12h2M4.6 19.4l1.4-1.4M18 6l1.4-1.4"
      />
    </svg>
  );
}

function SystemIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" className="h-4 w-4">
      <rect x="2.5" y="4" width="19" height="13" rx="2" />
      <path strokeLinecap="round" d="M8.5 20.5h7M12 17v3.5" />
    </svg>
  );
}

function MoonIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" className="h-4 w-4">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M20 14.2A8.2 8.2 0 0 1 9.8 4a8.2 8.2 0 1 0 10.2 10.2Z"
      />
    </svg>
  );
}

const OPTIONS: { value: Theme; labelKey: MessageKey; Icon: () => JSX.Element }[] = [
  { value: "light", labelKey: "theme.light", Icon: SunIcon },
  { value: "system", labelKey: "theme.system", Icon: SystemIcon },
  { value: "dark", labelKey: "theme.dark", Icon: MoonIcon },
];

export function ThemeToggle({ className = "" }: { className?: string }) {
  const { theme, setTheme } = useTheme();
  const { t } = useI18n();
  // 水合前不高亮，避免客户端读取本地存储后与 SSR 输出不一致
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);

  return (
    <div
      role="radiogroup"
      aria-label={t("theme.label")}
      className={`inline-flex items-center gap-0.5 rounded-lg border border-slate-200 bg-white p-0.5 dark:border-slate-700 dark:bg-slate-900 ${className}`}
    >
      {OPTIONS.map(({ value, labelKey, Icon }) => {
        const active = mounted && theme === value;
        const label = t(labelKey);
        return (
          <button
            key={value}
            type="button"
            role="radio"
            aria-checked={active}
            title={label}
            aria-label={label}
            onClick={() => setTheme(value)}
            className={`grid h-7 w-8 place-items-center rounded-md transition ${
              active
                ? "bg-indigo-600 text-white"
                : "text-slate-500 hover:bg-slate-100 hover:text-slate-700 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-200"
            }`}
          >
            <Icon />
          </button>
        );
      })}
    </div>
  );
}
