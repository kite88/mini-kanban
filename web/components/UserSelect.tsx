"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useI18n } from "@/components/I18nProvider";
import { inputCls } from "@/components/Modal";
import type { Department, User } from "@/lib/types";

interface Option {
  id: string | null;
  label: string;
  hint?: string;
  dept?: string;
}

/**
 * 可搜索的用户下拉：点击展开后用关键字过滤（匹配用户名、邮箱或所属部门），
 * 支持 ↑/↓ 选择、Enter 确认、Esc 关闭、点击外部关闭。
 */
export function UserSelect({
  users,
  departments = [],
  value,
  onChange,
  disabled,
  placeholder,
}: {
  users: User[];
  departments?: Department[];
  value: string | null;
  onChange: (id: string | null) => void;
  disabled?: boolean;
  placeholder?: string;
}) {
  const { t } = useI18n();
  // 未显式传入时用当前语言的默认提示语
  const placeholderText = placeholder ?? t("userSelect.placeholder");
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [active, setActive] = useState(0);
  const boxRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  const deptName = useMemo(() => {
    const map = new Map<string, string>();
    for (const d of departments) map.set(d.id, d.name);
    return (id: string | null) => (id == null ? "" : (map.get(id) ?? ""));
  }, [departments]);

  const selected = users.find((u) => u.id === value) ?? null;

  const options = useMemo<Option[]>(() => {
    const q = query.trim().toLowerCase();
    const matched = q
      ? users.filter(
          (u) =>
            u.username.toLowerCase().includes(q) ||
            u.email.toLowerCase().includes(q) ||
            deptName(u.departmentId).toLowerCase().includes(q),
        )
      : users;
    return [
      { id: null, label: t("common.unassigned") },
      ...matched.map((u) => ({
        id: u.id,
        label: u.username,
        hint: u.email,
        dept: deptName(u.departmentId),
      })),
    ];
  }, [users, query, deptName, t]);

  // 展开时重置搜索词并聚焦输入框
  useEffect(() => {
    if (!open) return;
    setQuery("");
    setActive(0);
    const id = requestAnimationFrame(() => searchRef.current?.focus());
    return () => cancelAnimationFrame(id);
  }, [open]);

  // 点击外部关闭
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, [open]);

  const commit = (id: string | null) => {
    onChange(id);
    setOpen(false);
  };

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActive((i) => Math.min(i + 1, options.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActive((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (options[active]) commit(options[active].id);
    } else if (e.key === "Escape") {
      // 只关下拉，不要冒泡到 Modal 去把整个弹窗也关掉
      e.preventDefault();
      e.stopPropagation();
      setOpen(false);
    }
  };

  return (
    <div ref={boxRef} className="relative">
      <button
        type="button"
        disabled={disabled}
        onClick={() => setOpen((v) => !v)}
        className={`${inputCls} flex items-center justify-between gap-2 text-left`}
      >
        <span className="flex min-w-0 items-center gap-1.5">
          <span
            className={`truncate ${selected ? "" : "text-slate-400 dark:text-slate-500"}`}
          >
            {selected ? selected.username : t("common.unassigned")}
          </span>
          {selected && deptName(selected.departmentId) && (
            <span className="shrink-0 rounded bg-slate-100 px-1 py-0.5 text-[10px] leading-none text-slate-500 dark:bg-slate-800 dark:text-slate-400">
              {deptName(selected.departmentId)}
            </span>
          )}
        </span>
        <span className="shrink-0 text-xs text-slate-400 dark:text-slate-500">▾</span>
      </button>

      {open && (
        <div className="absolute left-0 right-0 z-30 mt-1 rounded-lg border border-slate-200 bg-white p-1 shadow-xl dark:border-slate-700 dark:bg-slate-900">
          <input
            ref={searchRef}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setActive(0);
            }}
            onKeyDown={onKeyDown}
            placeholder={placeholderText}
            className="mb-1 w-full rounded-md border border-slate-200 bg-white px-2 py-1.5 text-sm text-slate-800 outline-none transition focus:border-indigo-500 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100 dark:placeholder:text-slate-500"
          />
          <ul className="max-h-56 overflow-y-auto">
            {options.length === 1 && (
              <li className="px-2 py-2 text-xs text-slate-400 dark:text-slate-500">
                {t("userSelect.noMatch")}
              </li>
            )}
            {options.map((o, i) => (
              <li key={o.id === null ? "none" : o.id}>
                <button
                  type="button"
                  onMouseEnter={() => setActive(i)}
                  onClick={() => commit(o.id)}
                  className={`flex w-full flex-col items-start gap-0.5 rounded-md px-2 py-1.5 text-left text-sm transition ${
                    i === active
                      ? "bg-indigo-50 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300"
                      : "text-slate-700 dark:text-slate-200"
                  }`}
                >
                  <span className="flex w-full min-w-0 items-center gap-1.5">
                    <span className="truncate">{o.label}</span>
                    {o.dept && (
                      <span className="shrink-0 rounded bg-slate-100 px-1 py-0.5 text-[10px] leading-none text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                        {o.dept}
                      </span>
                    )}
                  </span>
                  {o.hint && (
                    <span className="w-full truncate text-xs text-slate-400 dark:text-slate-500">
                      {o.hint}
                    </span>
                  )}
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
