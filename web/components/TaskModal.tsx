"use client";

import { useEffect, useRef, useState } from "react";
import { AttachmentList } from "@/components/AttachmentList";
import { useAuth } from "@/components/AuthProvider";
import { useI18n } from "@/components/I18nProvider";
import { Field, Modal, inputCls } from "@/components/Modal";
import { RichTextEditor, RichTextView } from "@/components/RichTextEditor";
import { UserSelect } from "@/components/UserSelect";
import { formatDateTime, fromInputValue, toInputValue } from "@/lib/date";
import {
  isRichTextEmpty,
  replaceFileTicket,
  resolveFileUrls,
  toStoredHtml,
} from "@/lib/html";
import { statusLabel } from "@/lib/labels";
import {
  COLUMNS,
  type Attachment,
  type Department,
  type Task,
  type TaskStatus,
  type User,
} from "@/lib/types";

export interface TaskFormValue {
  title: string;
  description: string;
  assigneeId: string | null;
  dueDate: string | null;
  status: TaskStatus;
  attachments: Attachment[];
}

// 只读信息展示框样式（预览模式复用）
const valueCls =
  "rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-300";

export function TaskModal({
  open,
  task,
  users,
  departments = [],
  defaultStatus,
  canEdit = true,
  onClose,
  onSubmit,
  onDelete,
}: {
  open: boolean;
  task: Task | null;
  users: User[];
  departments?: Department[]; // 用于在用户下拉里标注所属部门
  defaultStatus?: TaskStatus;
  canEdit?: boolean; // 是否有编辑权限（后端同样会校验）
  onClose: () => void;
  onSubmit: (v: TaskFormValue) => Promise<void>;
  onDelete?: () => void;
}) {
  const [title, setTitle] = useState(task?.title ?? "");
  // 富文本描述：展示 / 编辑时把 /api/files/xxx 补成后端地址，提交时再还原成相对地址
  const [description, setDescription] = useState(resolveFileUrls(task?.description ?? ""));
  const [attachments, setAttachments] = useState<Attachment[]>(task?.attachments ?? []);
  const [assigneeId, setAssigneeId] = useState<string | null>(task?.assigneeId ?? null);
  const [dueDate, setDueDate] = useState(toInputValue(task?.dueDate));
  const [status, setStatus] = useState<TaskStatus>(task?.status ?? defaultStatus ?? "todo");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  // 已有任务：先展示预览，点顶部「编辑」才切换成表单；新建任务直接就是表单
  const [editing, setEditing] = useState(!task);
  const pickerRef = useRef<HTMLInputElement>(null);
  const { user, fileTicket } = useAuth();
  const { t, locale } = useI18n();
  const isCreator = !!user && !!task && task.creatorId === user.id;
  const editable = canEdit && editing;

  // 下载票据首次取回或续期后，把描述里图片地址上的票据换成新的；
  // 描述里没有附件地址时字符串不变，不会触发重渲染。
  useEffect(() => {
    if (!fileTicket) return;
    setDescription((d) => replaceFileTicket(d, fileTicket));
  }, [fileTicket]);

  // 原生 datetime-local 的显示格式由浏览器语言决定，无法通过 lang / 属性改成中文，
  // 所以只把它当作「选择器」：界面上展示的是按站点语言格式化的只读输入框。
  const openPicker = () => {
    const el = pickerRef.current;
    if (!el) return;
    try {
      el.showPicker();
    } catch {
      el.focus();
    }
  };

  // 取消编辑：把表单恢复成任务当前的值，并回到预览
  const cancelEdit = () => {
    setTitle(task?.title ?? "");
    setDescription(resolveFileUrls(task?.description ?? ""));
    setAttachments(task?.attachments ?? []);
    setAssigneeId(task?.assigneeId ?? null);
    setDueDate(toInputValue(task?.dueDate));
    setStatus(task?.status ?? defaultStatus ?? "todo");
    setError("");
    setEditing(false);
  };

  // 弹窗每次打开时同步表单（key 由父组件传入保证状态重置）
  const submit = async () => {
    if (!title.trim()) {
      setError(t("task.titleRequired"));
      return;
    }
    setSaving(true);
    setError("");
    try {
      await onSubmit({
        title,
        description: isRichTextEmpty(description) ? "" : toStoredHtml(description),
        assigneeId,
        dueDate: fromInputValue(dueDate),
        status,
        attachments,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : t("common.saveFailed"));
    } finally {
      setSaving(false);
    }
  };

  const assigneeName =
    users.find((u) => u.id === assigneeId)?.username ??
    task?.assignee?.username ??
    t("common.unassigned");
  const statusTitle = statusLabel(t, status);
  const creatorName =
    task?.creator?.username ??
    (task?.creatorId ? t("task.unknownUser", { id: task.creatorId }) : t("common.unknown"));

  return (
    <Modal
      open={open}
      title={task ? (editable ? t("task.editTitle") : t("task.detailTitle")) : t("task.createTitle")}
      onClose={onClose}
      widthClass="max-w-3xl"
      headerExtra={
        task && canEdit && !editing ? (
          <button
            onClick={() => setEditing(true)}
            className="rounded-lg border border-slate-200 px-3 py-1 text-sm text-slate-600 transition hover:bg-slate-100 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {t("common.edit")}
          </button>
        ) : null
      }
      footer={
        <>
          {task && onDelete && (
            <button
              onClick={onDelete}
              className="mr-auto rounded-lg px-3 py-2 text-sm text-rose-600 transition hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-500/15"
            >
              {t("common.delete")}
            </button>
          )}
          {task && !onDelete && (
            <span className="mr-auto text-xs text-slate-400 dark:text-slate-500">
              {t("task.deleteNote")}
            </span>
          )}
          <button
            onClick={editable ? cancelEdit : onClose}
            className="rounded-lg border border-slate-200 px-4 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {editable ? t("common.cancel") : t("common.close")}
          </button>
          {editable && (
            <button
              onClick={submit}
              disabled={saving}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700 disabled:opacity-60"
            >
              {saving ? t("common.saving") : t("common.save")}
            </button>
          )}
        </>
      }
    >
      {!canEdit && task && (
        <p className="rounded-lg bg-slate-100 px-3 py-2 text-xs text-slate-500 dark:bg-slate-800 dark:text-slate-400">
          {t("task.readonlyNote")}
        </p>
      )}
      {error && (
        <p className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-600 dark:bg-rose-500/15 dark:text-rose-400">
          {error}
        </p>
      )}

      {editable ? (
        <>
          <Field label={t("task.field.title")}>
            <input
              className={inputCls}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={t("task.field.titlePlaceholder")}
              autoFocus={!task}
            />
          </Field>
          <Field label={t("task.field.description")}>
            <RichTextEditor
              value={description}
              onChange={setDescription}
              placeholder={t("task.field.descriptionPlaceholder")}
            />
          </Field>
          <Field label={t("task.field.attachments")}>
            <AttachmentList items={attachments} onChange={setAttachments} />
          </Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label={t("task.field.assignee")}>
              <UserSelect
                users={users}
                departments={departments}
                value={assigneeId}
                onChange={setAssigneeId}
              />
            </Field>
            <Field label={t("task.field.status")}>
              <select
                className={inputCls}
                value={status}
                onChange={(e) => setStatus(e.target.value as TaskStatus)}
              >
                {COLUMNS.map((c) => (
                  <option key={c.id} value={c.id}>
                    {statusLabel(t, c.id)}
                  </option>
                ))}
              </select>
            </Field>
          </div>
          <Field label={t("task.field.dueDate")}>
            <div className="relative">
              <input
                type="text"
                readOnly
                value={dueDate ? formatDateTime(dueDate, locale) : ""}
                placeholder={t("task.field.dueDatePlaceholder")}
                onClick={openPicker}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    openPicker();
                  }
                }}
                className={`${inputCls} cursor-pointer ${dueDate ? "pr-16" : "pr-10"}`}
              />
              {/* 透明覆盖层：保持渲染以便调用 showPicker，点击任意位置唤起原生选择器 */}
              <input
                ref={pickerRef}
                type="datetime-local"
                tabIndex={-1}
                aria-hidden="true"
                className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
                value={dueDate}
                onClick={openPicker}
                onChange={(e) => setDueDate(e.target.value)}
              />
              {dueDate ? (
                <button
                  type="button"
                  onClick={() => setDueDate("")}
                  className="absolute right-2 top-1/2 z-10 -translate-y-1/2 rounded px-1.5 py-0.5 text-xs text-slate-500 transition hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
                >
                  {t("common.clear")}
                </button>
              ) : (
                <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500">
                  🗓
                </span>
              )}
            </div>
          </Field>
          {/* 创建人是只读信息：新建任务时还没有，故只在编辑已有任务时展示 */}
          {task && (
            <Field label={t("task.field.creator")}>
              <p className={valueCls}>
                {creatorName}
                {isCreator && (
                  <span className="ml-1 text-xs text-indigo-600 dark:text-indigo-300">
                    {t("task.me")}
                  </span>
                )}
              </p>
            </Field>
          )}
        </>
      ) : (
        <>
          <Field label={t("task.field.title")}>
            <p className={`${valueCls} font-medium text-slate-800 dark:text-slate-100`}>{title}</p>
          </Field>
          <Field label={t("task.field.description")}>
            <RichTextView html={description} empty={t("task.noDescription")} />
          </Field>
          <Field label={t("task.field.attachments")}>
            <AttachmentList items={attachments} onChange={setAttachments} readOnly />
          </Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label={t("task.field.assignee")}>
              <p className={valueCls}>{assigneeName}</p>
            </Field>
            <Field label={t("task.field.status")}>
              <p className={valueCls}>{statusTitle}</p>
            </Field>
          </div>
          <Field label={t("task.field.dueDate")}>
            <p className={valueCls}>
              {dueDate ? formatDateTime(dueDate, locale) : t("common.notSet")}
            </p>
          </Field>
          <Field label={t("task.field.creator")}>
            <p className={valueCls}>
              {creatorName}
              {isCreator && (
                <span className="ml-1 text-xs text-indigo-600 dark:text-indigo-300">
                  {t("task.me")}
                </span>
              )}
            </p>
          </Field>
        </>
      )}
    </Modal>
  );
}
