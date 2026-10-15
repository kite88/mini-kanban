"use client";

import { useCallback, useEffect, useState } from "react";
import { AuthGuard } from "@/components/AuthGuard";
import { useConfirm } from "@/components/ConfirmProvider";
import { useI18n } from "@/components/I18nProvider";
import { api } from "@/lib/api";
import { formatDateTime } from "@/lib/date";
import { tNow } from "@/lib/i18n";
import { statusLabel } from "@/lib/labels";
import type {
  TrashDoc,
  TrashPayload,
  TrashProject,
  TrashTask,
} from "@/lib/types";

/** 回收箱：软删除的项目与任务集中在这里，可恢复或彻底删除 */
export default function TrashPage() {
  const { t, locale } = useI18n();
  const confirm = useConfirm();
  const [data, setData] = useState<TrashPayload>({ projects: [], tasks: [], docs: [] });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  // busy 记「正在操作的那一条」，用于禁用重复点击
  const [busy, setBusy] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      setData(await api.listTrash());
    } catch (err) {
      setError(err instanceof Error ? err.message : tNow("common.loadFailed"));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  // 统一处理「操作中禁用 → 失败提示 → 刷新列表」
  const run = async (key: string, action: () => Promise<unknown>) => {
    setBusy(key);
    setError("");
    try {
      await action();
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : tNow("common.operationFailed"));
    } finally {
      setBusy(null);
    }
  };

  const purgeProject = async (p: TrashProject) => {
    const parts = [
      p.taskCount > 0 ? t("trash.countTasks", { count: p.taskCount }) : "",
      p.docCount > 0 ? t("trash.countDocs", { count: p.docCount }) : "",
    ].filter(Boolean);
    const inner = parts.length
      ? t("trash.purgeProjectInner", { items: parts.join(t("common.listSeparator")) })
      : "";
    const ok = await confirm({
      title: t("trash.purgeProjectTitle"),
      description: t("trash.purgeProjectDesc", { title: p.title, inner }),
      note: t("trash.purgeProjectNote"),
      confirmText: t("trash.purge"),
      danger: true,
    });
    if (!ok) return;
    await run(`project-${p.id}`, () => api.purgeProject(p.id));
  };

  const purgeTask = async (task: TrashTask) => {
    const ok = await confirm({
      title: t("trash.purgeTaskTitle"),
      description: t("trash.purgeTaskDesc", { title: task.title }),
      note: t("trash.purgeNote"),
      confirmText: t("trash.purge"),
      danger: true,
    });
    if (!ok) return;
    await run(`task-${task.id}`, () => api.purgeTask(task.id));
  };

  const purgeDoc = async (d: TrashDoc) => {
    const ok = await confirm({
      title: t("trash.purgeDocTitle"),
      description: t("trash.purgeDocDesc", { title: d.title }),
      note: t("trash.purgeNote"),
      confirmText: t("trash.purge"),
      danger: true,
    });
    if (!ok) return;
    await run(`doc-${d.id}`, () => api.purgeDoc(d.id));
  };

  const empty = !data.projects.length && !data.tasks.length && !data.docs.length;

  return (
    <AuthGuard>
      <main className="mx-auto max-w-[1400px] px-6 py-8">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">
            {t("trash.title")}
          </h1>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{t("trash.subtitle")}</p>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-rose-50 px-4 py-2 text-sm text-rose-600 dark:bg-rose-500/15 dark:text-rose-400">
            {error}
          </div>
        )}

        {loading ? (
          <div className="py-10 text-center text-sm text-slate-500 dark:text-slate-400">
            {t("common.loading")}
          </div>
        ) : empty ? (
          <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-12 text-center text-slate-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-400">
            {t("trash.empty")}
          </div>
        ) : (
          <div className="space-y-8">
            <section>
              <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-200">
                {t("trash.projects")}
                <span className="rounded bg-slate-100 px-1.5 text-xs font-normal text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                  {data.projects.length}
                </span>
              </h2>
              {data.projects.length === 0 ? (
                <p className="rounded-xl bg-slate-50 px-4 py-3 text-xs text-slate-500 dark:bg-slate-900 dark:text-slate-400">
                  {t("trash.noProjects")}
                </p>
              ) : (
                <ul className="divide-y divide-slate-100 overflow-hidden rounded-2xl border border-slate-200 bg-white dark:divide-slate-800 dark:border-slate-800 dark:bg-slate-900">
                  {data.projects.map((p) => (
                    <li key={p.id} className="flex items-center justify-between gap-4 px-5 py-4">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-slate-800 dark:text-slate-100">
                          {p.title}
                        </p>
                        <p className="mt-1 truncate text-xs text-slate-400 dark:text-slate-500">
                          {p.owner?.username ? t("trash.projectSummaryCreator", { name: p.owner.username }) : ""}
                          {t("trash.projectSummary", {
                            tasks: p.taskCount,
                            docs: p.docCount,
                            time: formatDateTime(p.deletedAt, locale),
                          })}
                        </p>
                      </div>
                      <div className="flex shrink-0 gap-2">
                        <button
                          onClick={() => run(`project-${p.id}`, () => api.restoreProject(p.id))}
                          disabled={busy === `project-${p.id}`}
                          className="rounded-lg border border-slate-200 px-3 py-1.5 text-xs text-slate-600 transition hover:bg-slate-50 disabled:opacity-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
                        >
                          {t("trash.restore")}
                        </button>
                        <button
                          onClick={() => purgeProject(p)}
                          disabled={busy === `project-${p.id}`}
                          className="rounded-lg border border-rose-200 px-3 py-1.5 text-xs text-rose-600 transition hover:bg-rose-50 disabled:opacity-50 dark:border-rose-500/40 dark:text-rose-400 dark:hover:bg-rose-500/15"
                        >
                          {t("trash.purge")}
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section>
              <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-200">
                {t("trash.tasks")}
                <span className="rounded bg-slate-100 px-1.5 text-xs font-normal text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                  {data.tasks.length}
                </span>
              </h2>
              {data.tasks.length === 0 ? (
                <p className="rounded-xl bg-slate-50 px-4 py-3 text-xs text-slate-500 dark:bg-slate-900 dark:text-slate-400">
                  {t("trash.noTasks")}
                </p>
              ) : (
                <ul className="divide-y divide-slate-100 overflow-hidden rounded-2xl border border-slate-200 bg-white dark:divide-slate-800 dark:border-slate-800 dark:bg-slate-900">
                  {data.tasks.map((task) => (
                    <li key={task.id} className="flex items-center justify-between gap-4 px-5 py-4">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-slate-800 dark:text-slate-100">
                          {task.title}
                        </p>
                        <p className="mt-1 truncate text-xs text-slate-400 dark:text-slate-500">
                          {t("trash.taskSummary", {
                            project: task.projectTitle || t("trash.unknownProject"),
                            status: statusLabel(t, task.status),
                            time: formatDateTime(task.deletedAt, locale),
                          })}
                        </p>
                      </div>
                      <div className="flex shrink-0 gap-2">
                        <button
                          onClick={() => run(`task-${task.id}`, () => api.restoreTask(task.id))}
                          disabled={busy === `task-${task.id}`}
                          className="rounded-lg border border-slate-200 px-3 py-1.5 text-xs text-slate-600 transition hover:bg-slate-50 disabled:opacity-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
                        >
                          {t("trash.restore")}
                        </button>
                        <button
                          onClick={() => purgeTask(task)}
                          disabled={busy === `task-${task.id}`}
                          className="rounded-lg border border-rose-200 px-3 py-1.5 text-xs text-rose-600 transition hover:bg-rose-50 disabled:opacity-50 dark:border-rose-500/40 dark:text-rose-400 dark:hover:bg-rose-500/15"
                        >
                          {t("trash.purge")}
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section>
              <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-200">
                {t("trash.docs")}
                <span className="rounded bg-slate-100 px-1.5 text-xs font-normal text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                  {data.docs.length}
                </span>
              </h2>
              {data.docs.length === 0 ? (
                <p className="rounded-xl bg-slate-50 px-4 py-3 text-xs text-slate-500 dark:bg-slate-900 dark:text-slate-400">
                  {t("trash.noDocs")}
                </p>
              ) : (
                <ul className="divide-y divide-slate-100 overflow-hidden rounded-2xl border border-slate-200 bg-white dark:divide-slate-800 dark:border-slate-800 dark:bg-slate-900">
                  {data.docs.map((d) => (
                    <li key={d.id} className="flex items-center justify-between gap-4 px-5 py-4">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-slate-800 dark:text-slate-100">
                          {d.title}
                        </p>
                        <p className="mt-1 truncate text-xs text-slate-400 dark:text-slate-500">
                          {t("trash.docSummary", {
                            project: d.projectTitle || t("trash.unknownProject"),
                            time: formatDateTime(d.deletedAt, locale),
                          })}
                          {!!d.attachments?.length &&
                            t("trash.attachmentSuffix", { count: d.attachments.length })}
                        </p>
                      </div>
                      <div className="flex shrink-0 gap-2">
                        <button
                          onClick={() => run(`doc-${d.id}`, () => api.restoreDoc(d.id))}
                          disabled={busy === `doc-${d.id}`}
                          className="rounded-lg border border-slate-200 px-3 py-1.5 text-xs text-slate-600 transition hover:bg-slate-50 disabled:opacity-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
                        >
                          {t("trash.restore")}
                        </button>
                        <button
                          onClick={() => purgeDoc(d)}
                          disabled={busy === `doc-${d.id}`}
                          className="rounded-lg border border-rose-200 px-3 py-1.5 text-xs text-rose-600 transition hover:bg-rose-50 disabled:opacity-50 dark:border-rose-500/40 dark:text-rose-400 dark:hover:bg-rose-500/15"
                        >
                          {t("trash.purge")}
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </div>
        )}
      </main>
    </AuthGuard>
  );
}
