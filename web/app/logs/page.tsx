"use client";

import { useCallback, useEffect, useState } from "react";
import { AuthGuard } from "@/components/AuthGuard";
import { useI18n } from "@/components/I18nProvider";
import { Modal, inputCls } from "@/components/Modal";
import { api } from "@/lib/api";
import { formatDateTime } from "@/lib/date";
import { tNow } from "@/lib/i18n";
import { LOG_ACTIONS, LOG_TARGET_TYPES, logActionLabel, logTargetLabel } from "@/lib/labels";
import type { LogAction, OperationLog } from "@/lib/types";

const PAGE_SIZE = 20;

/** 动作徽标配色：物理删除 / 删除用警示色，新建 / 恢复用正向色 */
const ACTION_CLS: Record<LogAction, string> = {
  create: "bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300",
  update: "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300",
  move: "bg-blue-50 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300",
  delete: "bg-amber-50 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300",
  restore: "bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300",
  purge: "bg-rose-50 text-rose-600 dark:bg-rose-500/15 dark:text-rose-300",
};

/** 操作日志：项目 / 任务 / 项目资料 / 用户 / 部门的操作留痕，物理删除的条目可查看数据快照 */
export default function LogsPage() {
  const { t, locale } = useI18n();
  const [items, setItems] = useState<OperationLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [targetType, setTargetType] = useState("");
  const [action, setAction] = useState("");
  // keyword 是真正生效的筛选值，keywordInput 是输入框内容（回车 / 点查询才生效）
  const [keyword, setKeyword] = useState("");
  const [keywordInput, setKeywordInput] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [detail, setDetail] = useState<OperationLog | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.listLogs({ targetType, action, keyword, page, pageSize: PAGE_SIZE });
      setItems(data.items);
      setTotal(data.total);
      setError("");
    } catch (err) {
      setError(err instanceof Error ? err.message : tNow("common.loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [targetType, action, keyword, page]);

  useEffect(() => {
    load();
  }, [load]);

  // 筛选条件变化一律回到第一页，避免停在越界的页码上
  const applyFilter = (fn: () => void) => {
    fn();
    setPage(1);
  };

  const openDetail = async (row: OperationLog) => {
    setDetail(null);
    setDetailOpen(true);
    setDetailLoading(true);
    try {
      setDetail(await api.getLog(row.id));
    } catch (err) {
      setDetailOpen(false);
      setError(err instanceof Error ? err.message : tNow("logs.detailFailed"));
    } finally {
      setDetailLoading(false);
    }
  };

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <AuthGuard>
      <main className="mx-auto max-w-[1400px] px-6 py-8">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">
            {t("logs.title")}
          </h1>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{t("logs.subtitle")}</p>
        </div>

        <div className="mb-4 flex flex-wrap items-center gap-2">
          <select
            className={`${inputCls} w-auto`}
            value={targetType}
            onChange={(e) => applyFilter(() => setTargetType(e.target.value))}
          >
            <option value="">{t("logs.filter.allTargets")}</option>
            {LOG_TARGET_TYPES.map((value) => (
              <option key={value} value={value}>
                {logTargetLabel(t, value)}
              </option>
            ))}
          </select>
          <select
            className={`${inputCls} w-auto`}
            value={action}
            onChange={(e) => applyFilter(() => setAction(e.target.value))}
          >
            <option value="">{t("logs.filter.allActions")}</option>
            {LOG_ACTIONS.map((value) => (
              <option key={value} value={value}>
                {logActionLabel(t, value)}
              </option>
            ))}
          </select>
          <form
            className="flex shrink-0 items-center gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              applyFilter(() => setKeyword(keywordInput.trim()));
            }}
          >
            <input
              className={`${inputCls} w-64 shrink-0`}
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              placeholder={t("logs.searchPlaceholder")}
            />
            <button
              type="submit"
              className="min-w-[4.5rem] whitespace-nowrap rounded-lg border border-slate-200 px-4 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {t("common.query")}
            </button>
            {(targetType || action || keyword) && (
              <button
                type="button"
                onClick={() => {
                  setKeywordInput("");
                  applyFilter(() => {
                    setTargetType("");
                    setAction("");
                    setKeyword("");
                  });
                }}
                className="min-w-[4.5rem] whitespace-nowrap rounded-lg px-4 py-2 text-sm text-slate-500 transition hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
              >
                {t("common.reset")}
              </button>
            )}
          </form>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-rose-50 px-4 py-2 text-sm text-rose-600 dark:bg-rose-500/15 dark:text-rose-400">
            {error}
          </div>
        )}

        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-card dark:border-slate-800 dark:bg-slate-900">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-slate-500 dark:bg-slate-950/60 dark:text-slate-400">
              <tr>
                <th className="px-4 py-3 font-medium">{t("logs.col.time")}</th>
                <th className="px-4 py-3 font-medium">{t("logs.col.operator")}</th>
                <th className="px-4 py-3 font-medium">{t("logs.col.targetType")}</th>
                <th className="px-4 py-3 font-medium">{t("logs.col.action")}</th>
                <th className="px-4 py-3 font-medium">{t("logs.col.target")}</th>
                <th className="px-4 py-3 font-medium">{t("logs.col.summary")}</th>
                <th className="px-4 py-3 text-right font-medium">{t("logs.col.detail")}</th>
              </tr>
            </thead>
            <tbody>
              {items.map((row) => (
                <tr key={row.id} className="border-t border-slate-100 dark:border-slate-800">
                  <td className="whitespace-nowrap px-4 py-3 text-slate-600 dark:text-slate-300">
                    {formatDateTime(row.createdAt, locale)}
                  </td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                    {row.operatorName || t("common.empty")}
                  </td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                    {logTargetLabel(t, row.targetType)}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded px-2 py-0.5 text-xs ${
                        ACTION_CLS[row.action] ??
                        "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300"
                      }`}
                    >
                      {logActionLabel(t, row.action)}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-medium text-slate-800 dark:text-slate-100">
                    {row.targetName || `#${row.targetId}`}
                  </td>
                  <td className="px-4 py-3 text-slate-500 dark:text-slate-400">
                    {row.summary || t("common.empty")}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openDetail(row)}
                      className="rounded px-2 py-1 text-slate-500 transition hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
                    >
                      {t("common.view")}
                    </button>
                  </td>
                </tr>
              ))}
              {!loading && items.length === 0 && (
                <tr>
                  <td
                    colSpan={7}
                    className="px-4 py-10 text-center text-slate-400 dark:text-slate-500"
                  >
                    {t("logs.empty")}
                  </td>
                </tr>
              )}
              {loading && (
                <tr>
                  <td
                    colSpan={7}
                    className="px-4 py-10 text-center text-slate-400 dark:text-slate-500"
                  >
                    {t("common.loading")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div className="mt-4 flex items-center justify-between text-sm text-slate-500 dark:text-slate-400">
          <span>
            {t("logs.count", { total, page, pages: totalPages })}
          </span>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1 || loading}
              className="rounded-lg border border-slate-200 px-3 py-1.5 text-xs text-slate-600 transition hover:bg-slate-50 disabled:opacity-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {t("logs.prev")}
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages || loading}
              className="rounded-lg border border-slate-200 px-3 py-1.5 text-xs text-slate-600 transition hover:bg-slate-50 disabled:opacity-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {t("logs.next")}
            </button>
          </div>
        </div>

        <Modal
          open={detailOpen}
          title={t("logs.detailTitle")}
          widthClass="max-w-3xl"
          onClose={() => setDetailOpen(false)}
        >
          {detailLoading || !detail ? (
            <p className="py-6 text-center text-sm text-slate-500 dark:text-slate-400">
              {t("common.loading")}
            </p>
          ) : (
            <>
              <dl className="grid grid-cols-[6rem_1fr] gap-x-4 gap-y-2 text-sm">
                <dt className="text-slate-500 dark:text-slate-400">{t("logs.col.time")}</dt>
                <dd className="text-slate-800 dark:text-slate-100">
                  {formatDateTime(detail.createdAt, locale)}
                </dd>
                <dt className="text-slate-500 dark:text-slate-400">{t("logs.col.operator")}</dt>
                <dd className="text-slate-800 dark:text-slate-100">
                  {detail.operatorName || t("common.empty")}
                  <span className="ml-2 text-xs text-slate-400 dark:text-slate-500">
                    #{detail.operatorId}
                  </span>
                </dd>
                <dt className="text-slate-500 dark:text-slate-400">{t("logs.col.target")}</dt>
                <dd className="text-slate-800 dark:text-slate-100">
                  {logTargetLabel(t, detail.targetType)} ·{" "}
                  {detail.targetName || t("common.empty")}
                  <span className="ml-2 text-xs text-slate-400 dark:text-slate-500">
                    #{detail.targetId}
                  </span>
                </dd>
                <dt className="text-slate-500 dark:text-slate-400">{t("logs.col.action")}</dt>
                <dd>
                  <span
                    className={`rounded px-2 py-0.5 text-xs ${
                      ACTION_CLS[detail.action] ??
                      "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300"
                    }`}
                  >
                    {logActionLabel(t, detail.action)}
                  </span>
                </dd>
                <dt className="text-slate-500 dark:text-slate-400">{t("logs.col.summary")}</dt>
                <dd className="text-slate-800 dark:text-slate-100">
                  {detail.summary || t("common.empty")}
                </dd>
              </dl>

              <div>
                <p className="mb-2 text-sm text-slate-500 dark:text-slate-400">{t("logs.snapshot")}</p>
                {detail.snapshot ? (
                  <pre className="max-h-96 overflow-auto rounded-lg bg-slate-50 p-3 text-xs leading-relaxed text-slate-700 dark:bg-slate-950 dark:text-slate-300">
                    {JSON.stringify(detail.snapshot, null, 2)}
                  </pre>
                ) : (
                  <p className="rounded-lg bg-slate-50 px-4 py-3 text-xs text-slate-500 dark:bg-slate-950 dark:text-slate-400">
                    {t("logs.snapshotEmpty")}
                  </p>
                )}
              </div>
            </>
          )}
        </Modal>
      </main>
    </AuthGuard>
  );
}
