"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { useI18n } from "@/components/I18nProvider";

export interface ConfirmOptions {
  /** 标题：一句话说清要做什么，如「删除任务」 */
  title: string;
  /** 正文：待确认的具体对象，通常带上对象名称 */
  description?: React.ReactNode;
  /** 结果提示：单独一条横幅，交代「能不能恢复」这类关键后果 */
  note?: string;
  /** 确认按钮文案，默认「确认」 */
  confirmText?: string;
  /** 取消按钮文案，默认「取消」 */
  cancelText?: string;
  /**
   * 危险操作：确认按钮变红，并把默认焦点放在「取消」上 ——
   * 回车会先落到取消按钮，避免顺手一敲就把数据删了。
   */
  danger?: boolean;
}

type ConfirmFn = (options: ConfirmOptions) => Promise<boolean>;

const ConfirmContext = createContext<ConfirmFn | null>(null);

const overlayCls =
  "fixed inset-0 z-[60] grid animate-confirm-overlay place-items-center bg-slate-900/45 p-4 backdrop-blur-[2px]";

const panelCls =
  "w-full max-w-sm animate-confirm-dialog overflow-hidden rounded-2xl bg-white shadow-2xl ring-1 ring-slate-900/5 motion-reduce:animate-none dark:bg-slate-900 dark:ring-slate-800";

const iconDangerCls = "bg-rose-100 text-rose-600 dark:bg-rose-500/15 dark:text-rose-400";
const iconNormalCls = "bg-indigo-100 text-indigo-600 dark:bg-indigo-500/15 dark:text-indigo-300";
const noteDangerCls = "bg-rose-50 text-rose-600 dark:bg-rose-500/10 dark:text-rose-300";
const noteNormalCls = "bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300";

// focus: 而不是 focus-visible:：弹窗打开时按钮是程序化聚焦的，
// 用 focus-visible 在鼠标操作后不显示焦点环，键盘用户会看不出焦点在哪
const btnBase =
  "rounded-lg px-4 py-2 text-sm font-medium transition focus:outline-none focus:ring-2 focus:ring-offset-0 disabled:opacity-60";
const cancelBtnCls = `${btnBase} border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 focus:ring-slate-400/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800`;
const primaryBtnCls = `${btnBase} bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500/50`;
const dangerBtnCls = `${btnBase} bg-rose-600 text-white hover:bg-rose-700 focus:ring-rose-500/50`;

/**
 * ConfirmProvider 全局确认框：把原生 window.confirm 换成项目自己的弹窗。
 *
 * 原生 confirm 的两个问题：样式完全不受控（各浏览器自成一派，与站内主题不搭），
 * 且它是同步阻塞的，弹窗期间整个渲染线程停摆。这里给出的是一个 Promise 版本：
 *
 *   const confirm = useConfirm();
 *   if (!(await confirm({ title: "删除任务", description: `确认删除「${t.title}」？`, danger: true }))) return;
 *
 * 弹窗挂在 body 层级、z-index 高于 Modal，因此在弹窗里再次确认也不会被遮住。
 */
export function ConfirmProvider({ children }: { children: React.ReactNode }) {
  const { t } = useI18n();
  const [options, setOptions] = useState<ConfirmOptions | null>(null);
  const resolveRef = useRef<((ok: boolean) => void) | null>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);
  const okRef = useRef<HTMLButtonElement>(null);

  const confirm = useCallback<ConfirmFn>((opts) => {
    // 上一个还没关闭就被再次调用：先把它按「取消」收尾，避免那个 Promise 永远悬着
    resolveRef.current?.(false);
    return new Promise<boolean>((resolve) => {
      resolveRef.current = resolve;
      setOptions(opts);
    });
  }, []);

  const close = useCallback((ok: boolean) => {
    const resolve = resolveRef.current;
    resolveRef.current = null;
    setOptions(null);
    resolve?.(ok);
  }, []);

  // 打开时把焦点放进弹窗：危险操作落在「取消」上
  useEffect(() => {
    if (!options) return;
    (options.danger ? cancelRef : okRef).current?.focus();
  }, [options]);

  // 键盘处理挂在 window 上而不是弹窗节点上：点过弹窗里的正文区后焦点会回到 body，
  // 这时挂在节点上的 onKeyDown 收不到事件，Esc 就失灵了
  useEffect(() => {
    if (!options) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        close(false);
        return;
      }
      if (e.key !== "Tab") return;
      const panel = panelRef.current;
      if (!panel) return;
      const buttons = panel.querySelectorAll<HTMLElement>("button");
      if (buttons.length === 0) return;
      const first = buttons[0];
      const last = buttons[buttons.length - 1];
      // 焦点已经跑到弹窗外（或落在 body 上）：先拉回弹窗里
      if (!panel.contains(document.activeElement)) {
        e.preventDefault();
        first.focus();
        return;
      }
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [options, close]);

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      {options && (
        <div
          className={overlayCls}
          // 点遮罩等于取消：只在按下的目标就是遮罩本身时生效，
          // 避免从弹窗内按住拖到外面时误触发
          onMouseDown={(e) => {
            if (e.target === e.currentTarget) close(false);
          }}
        >
          <div
            ref={panelRef}
            role="alertdialog"
            aria-modal="true"
            aria-labelledby="confirm-title"
            aria-describedby={options.description ? "confirm-desc" : undefined}
            className={panelCls}
          >
            <div className="flex gap-4 px-6 pb-5 pt-6">
              <span
                className={`grid h-10 w-10 shrink-0 place-items-center rounded-full ${
                  options.danger ? iconDangerCls : iconNormalCls
                }`}
              >
                {options.danger ? <AlertIcon /> : <QuestionIcon />}
              </span>
              <div className="min-w-0 flex-1">
                <h2
                  id="confirm-title"
                  className="text-base font-semibold leading-6 text-slate-900 dark:text-slate-100"
                >
                  {options.title}
                </h2>
                {options.description && (
                  <p
                    id="confirm-desc"
                    className="mt-1.5 break-words text-sm leading-relaxed text-slate-600 dark:text-slate-300"
                  >
                    {options.description}
                  </p>
                )}
                {options.note && (
                  <p
                    className={`mt-2.5 rounded-lg px-2.5 py-2 text-xs leading-relaxed ${
                      options.danger ? noteDangerCls : noteNormalCls
                    }`}
                  >
                    {options.note}
                  </p>
                )}
              </div>
            </div>
            <div className="flex justify-end gap-2 border-t border-slate-100 bg-slate-50/80 px-4 py-3 dark:border-slate-800 dark:bg-slate-950/40">
              <button ref={cancelRef} type="button" onClick={() => close(false)} className={cancelBtnCls}>
                {options.cancelText ?? t("common.cancel")}
              </button>
              <button
                ref={okRef}
                type="button"
                onClick={() => close(true)}
                className={options.danger ? dangerBtnCls : primaryBtnCls}
              >
                {options.confirmText ?? t("common.confirm")}
              </button>
            </div>
          </div>
        </div>
      )}
    </ConfirmContext.Provider>
  );
}

/** useConfirm 取全局确认框：返回一个 await 得到 boolean 的 confirm(fn) */
export function useConfirm(): ConfirmFn {
  const ctx = useContext(ConfirmContext);
  if (!ctx) throw new Error("useConfirm 必须在 ConfirmProvider 内使用");
  return ctx;
}

function AlertIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-5 w-5"
      aria-hidden="true"
    >
      <path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z" />
      <path d="M12 9v4" />
      <path d="M12 17h.01" />
    </svg>
  );
}

function QuestionIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-5 w-5"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="10" />
      <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
      <path d="M12 17h.01" />
    </svg>
  );
}
