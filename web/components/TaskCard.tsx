"use client";

import { memo } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useAuth } from "@/components/AuthProvider";
import { useI18n } from "@/components/I18nProvider";
import { formatDateTime, isOverdue } from "@/lib/date";
import { richTextToText } from "@/lib/html";
import type { Task } from "@/lib/types";

function initials(name?: string) {
  if (!name) return "?";
  return name.slice(0, 1).toUpperCase();
}

export function TaskCardBody({ task, dragging }: { task: Task; dragging?: boolean }) {
  const { user } = useAuth();
  const { t, locale } = useI18n();
  // 当前登录用户就是任务创建人时，在标题前打个标记
  const mine = !!user && task.creatorId === user.id;
  // 已完成 / 已验收都是终态，不再提示逾期
  const overdue = isOverdue(task.dueDate) && task.status !== "done" && task.status !== "accepted";
  return (
    <div
      className={`rounded-xl border bg-white p-3 shadow-sm transition dark:bg-slate-900 ${
        dragging
          ? "border-indigo-400 opacity-90 shadow-lg dark:border-indigo-500"
          : "border-slate-200 hover:border-indigo-300 dark:border-slate-700 dark:hover:border-indigo-500/60"
      }`}
    >
      <div className="flex items-start justify-between gap-2">
        <p className="text-sm font-medium leading-5 text-slate-800 dark:text-slate-100">
          {mine && (
            <span className="mr-1.5 inline-block rounded bg-indigo-50 px-1 py-0.5 align-middle text-[10px] font-normal leading-none text-indigo-600 dark:bg-indigo-500/20 dark:text-indigo-300">
              {t("task.createdByMe")}
            </span>
          )}
          {task.title}
        </p>
        {task.assignee && (
          <span
            title={task.assignee.username}
            className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-indigo-100 text-[11px] font-semibold text-indigo-700 dark:bg-indigo-500/20 dark:text-indigo-300"
          >
            {initials(task.assignee.username)}
          </span>
        )}
      </div>

      {task.description && (
        <p className="mt-1.5 line-clamp-2 text-xs text-slate-500 dark:text-slate-400">
          {richTextToText(task.description)}
        </p>
      )}

      <div className="mt-3 flex items-center justify-between text-[11px] text-slate-400 dark:text-slate-500">
        <span className="truncate">
          {task.assignee?.username ?? t("common.unassigned")}
          {!!task.attachments?.length && (
            <span
              className="ml-1.5"
              title={t("task.attachmentCount", { count: task.attachments.length })}
            >
              📎 {task.attachments.length}
            </span>
          )}
        </span>
        {task.dueDate && (
          <span
            className={`rounded px-1.5 py-0.5 ${
              overdue
                ? "bg-rose-50 text-rose-600 dark:bg-rose-500/15 dark:text-rose-400"
                : "bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400"
            }`}
          >
            {overdue
              ? t("task.overdueAt", { time: formatDateTime(task.dueDate, locale) })
              : t("task.dueAt", { time: formatDateTime(task.dueDate, locale) })}
          </span>
        )}
      </div>
    </div>
  );
}

// memo 化：拖拽期间看板会频繁重渲染，卡片不必跟着重渲染，
// 否则可拖拽项的测量（dnd-kit useRect）更容易和渲染互相触发。
export const TaskCard = memo(function TaskCard({
  task,
  onOpen,
  disabled,
}: {
  task: Task;
  onOpen: (t: Task) => void;
  disabled?: boolean; // 无编辑权限时禁用拖拽（后端 MoveTask 同样会校验）
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: String(task.id),
    disabled,
  });

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Translate.toString(transform), transition }}
      className={`${disabled ? "" : "touch-none"} ${isDragging ? "opacity-40" : ""}`}
      {...attributes}
      {...listeners}
      onClick={() => onOpen(task)}
    >
      <TaskCardBody task={task} />
    </div>
  );
});
