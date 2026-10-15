"use client";

import { useEffect } from "react";

interface ModalProps {
  open: boolean;
  title: string;
  onClose: () => void;
  children: React.ReactNode;
  footer?: React.ReactNode;
  /** 弹窗最大宽度（Tailwind 类），默认 max-w-md；内容多的弹窗可传更宽的 */
  widthClass?: string;
  /** 标题行右侧的额外操作（如「编辑」按钮），显示在关闭按钮左边 */
  headerExtra?: React.ReactNode;
  /**
   * 固定高度（Tailwind 高度类，如 h-[80vh]）：标题行与底部按钮固定，只有内容区滚动。
   * 不传时高度随内容自适应，超过 max-h-[90vh] 才整体滚动。
   */
  heightClass?: string;
}

export function Modal({
  open,
  title,
  onClose,
  children,
  footer,
  widthClass = "max-w-md",
  headerExtra,
  heightClass,
}: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  if (!open) return null;

  const fixed = !!heightClass;

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-slate-900/40 p-4 dark:bg-slate-950/70">
      <div
        className={`w-full rounded-2xl bg-white p-6 shadow-xl dark:bg-slate-900 dark:ring-1 dark:ring-slate-800 ${widthClass} ${
          fixed ? `flex flex-col ${heightClass}` : "max-h-[90vh] overflow-y-auto"
        }`}
      >
        <div className={`mb-4 flex items-center justify-between gap-3 ${fixed ? "shrink-0" : ""}`}>
          <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">{title}</h2>
          <div className="flex shrink-0 items-center gap-2">
            {headerExtra}
            <button
              onClick={onClose}
              className="rounded-lg px-2 py-1 text-slate-400 transition hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-800 dark:hover:text-slate-200"
            >
              ✕
            </button>
          </div>
        </div>
        <div className={fixed ? "min-h-0 flex-1 space-y-4 overflow-y-auto" : "space-y-4"}>
          {children}
        </div>
        {footer && (
          <div
            className={`flex shrink-0 justify-end gap-2 ${fixed ? "mt-4" : "mt-6"}`}
          >
            {footer}
          </div>
        )}
      </div>
    </div>
  );
}

export function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="mb-1 block text-sm text-slate-600 dark:text-slate-300">{label}</label>
      {children}
    </div>
  );
}

export const inputCls =
  "w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100 disabled:cursor-not-allowed disabled:opacity-60 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100 dark:placeholder:text-slate-500 dark:focus:border-indigo-400 dark:focus:ring-indigo-500/20";
