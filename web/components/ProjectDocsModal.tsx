"use client";

import { useCallback, useEffect, useState } from "react";
import { AttachmentList } from "@/components/AttachmentList";
import { useAuth } from "@/components/AuthProvider";
import { useConfirm } from "@/components/ConfirmProvider";
import { useI18n } from "@/components/I18nProvider";
import { Field, Modal, inputCls } from "@/components/Modal";
import { RichTextEditor, RichTextView } from "@/components/RichTextEditor";
import { api } from "@/lib/api";
import { formatDateTime } from "@/lib/date";
import { tNow } from "@/lib/i18n";
import {
  isRichTextEmpty,
  replaceFileTicket,
  resolveFileUrls,
  richTextToText,
  toStoredHtml,
} from "@/lib/html";
import type { Attachment, Project, ProjectDoc } from "@/lib/types";

// 只读信息展示框样式（与 TaskModal 保持一致）
const valueCls =
  "rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-300";

/**
 * 项目资料：存放需求文档 / 会议纪要 / 设计稿 / 项目密钥 / 测试账密等项目相关内容。
 *
 * 弹窗内两种视图：列表（浏览 + 新建入口）与单条资料（详情 / 编辑 / 新建表单）。
 * 表单只保留「标题 + 内容 + 附件」，相比新建任务去掉了关联人 / 所属列 / 任务到期时间。
 */
export function ProjectDocsModal({
  open,
  project,
  onClose,
}: {
  open: boolean;
  /** 所属项目：用于判断新建与编辑权限（与看板页同一套规则） */
  project: Project;
  onClose: () => void;
}) {
  const { user, fileTicket } = useAuth();
  const { t, locale } = useI18n();
  const confirm = useConfirm();
  const [docs, setDocs] = useState<ProjectDoc[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // 当前查看 / 编辑的资料；null 表示「还没有选中」或正在新建
  const [current, setCurrent] = useState<ProjectDoc | null>(null);
  // 视图：list 列表；doc 单条资料
  const [view, setView] = useState<"list" | "doc">("list");
  // 单条资料是否处于表单态（新建资料恒为表单）
  const [editing, setEditing] = useState(false);

  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState("");
  // 按标题检索（资料多了以后用）：只在前端过滤已加载的列表，输入即生效
  const [keyword, setKeyword] = useState("");

  const isAdmin = user?.role === "admin";
  const isOwner = !!user && project.ownerId === user.id;
  // 是否仍是项目成员（后端 isProjectMember 同一规则：关联表里的成员，
  // 外加项目创建人——创建人始终在关联表里，这里补一道防御）
  const isMember =
    isAdmin || (!!user && ((project.members ?? []).some((m) => m.id === user.id) || isOwner));
  // 新建资料权限：项目成员
  const canCreate = isMember;
  // 编辑 / 删除权限：必须是项目成员，且是管理员、资料创建人或项目创建人
  // （被移出项目后不再能处置自己的旧资料；后端 canManageDoc 同一规则）
  const canManage = (d: ProjectDoc) =>
    isMember && (isAdmin || (!!user && d.creatorId === user.id) || isOwner);

  // 标题检索：忽略大小写；关键词为空时展示全部
  const query = keyword.trim().toLowerCase();
  const visibleDocs = query ? docs.filter((d) => d.title.toLowerCase().includes(query)) : docs;

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setDocs(await api.listProjectDocs(project.id));
      setError("");
    } catch (err) {
      // 非渲染路径（异步回调）用 tNow：避免把 t 放进依赖后，切语言就整屏重新加载
      setError(err instanceof Error ? err.message : tNow("common.loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [project.id]);

  useEffect(() => {
    load();
  }, [load]);

  // 下载票据取回 / 续期后，刷新正文里图片地址上的票据
  useEffect(() => {
    if (!fileTicket) return;
    setContent((c) => replaceFileTicket(c, fileTicket));
  }, [fileTicket]);

  const backToList = async () => {
    setView("list");
    setCurrent(null);
    setEditing(false);
    setFormError("");
    await load();
  };

  const openDoc = (d: ProjectDoc) => {
    setCurrent(d);
    setView("doc");
    setEditing(false);
    setTitle(d.title);
    setContent(resolveFileUrls(d.content));
    setAttachments(d.attachments ?? []);
    setFormError("");
  };

  const startCreate = () => {
    setCurrent(null);
    setView("doc");
    setEditing(true);
    setTitle("");
    setContent("");
    setAttachments([]);
    setFormError("");
  };

  // 取消编辑：把表单恢复成资料当前的值，并回到详情
  const cancelEdit = () => {
    if (!current) return;
    setTitle(current.title);
    setContent(resolveFileUrls(current.content));
    setAttachments(current.attachments ?? []);
    setFormError("");
    setEditing(false);
  };

  const submit = async () => {
    if (!title.trim()) {
      setFormError(t("docs.titleRequired"));
      return;
    }
    setSaving(true);
    setFormError("");
    const payload = {
      title,
      content: isRichTextEmpty(content) ? "" : toStoredHtml(content),
      attachments,
    };
    try {
      if (current) await api.updateProjectDoc(current.id, payload);
      else await api.createProjectDoc(project.id, payload);
      await backToList();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : t("common.saveFailed"));
    } finally {
      setSaving(false);
    }
  };

  const remove = async () => {
    if (!current) return;
    const ok = await confirm({
      title: t("docs.deleteTitle"),
      description: t("docs.deleteDesc", { title: current.title }),
      note: t("docs.deleteNote"),
      confirmText: t("common.delete"),
      danger: true,
    });
    if (!ok) return;
    try {
      await api.deleteProjectDoc(current.id);
      await backToList();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : t("common.deleteFailed"));
    }
  };

  const docTitle = !current
    ? t("docs.createTitle")
    : editing
      ? t("docs.editTitle")
      : t("docs.detailTitle");
  const modalTitle = view === "list" ? t("docs.title") : docTitle;

  return (
    <Modal
      open={open}
      title={modalTitle}
      onClose={onClose}
      widthClass="max-w-3xl"
      heightClass="h-[80vh]"
      headerExtra={
        view === "list" ? (
          canCreate ? (
            <button
              onClick={startCreate}
              className="rounded-lg bg-indigo-600 px-3 py-1 text-sm font-medium text-white transition hover:bg-indigo-700"
            >
              {t("docs.create")}
            </button>
          ) : null
        ) : current && !editing && canManage(current) ? (
          <button
            onClick={() => setEditing(true)}
            className="rounded-lg border border-slate-200 px-3 py-1 text-sm text-slate-600 transition hover:bg-slate-100 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {t("common.edit")}
          </button>
        ) : null
      }
      footer={
        view === "list" ? (
          <button
            onClick={onClose}
            className="rounded-lg border border-slate-200 px-4 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {t("common.close")}
          </button>
        ) : (
          <>
            {current && canManage(current) && (
              <button
                onClick={remove}
                className="mr-auto rounded-lg px-3 py-2 text-sm text-rose-600 transition hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-500/15"
              >
                {t("common.delete")}
              </button>
            )}
            <button
              onClick={editing ? (current ? cancelEdit : backToList) : backToList}
              className="rounded-lg border border-slate-200 px-4 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {editing ? t("common.cancel") : t("docs.backToList")}
            </button>
            {editing && (
              <button
                onClick={submit}
                disabled={saving}
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700 disabled:opacity-60"
              >
                {saving ? t("common.saving") : t("common.save")}
              </button>
            )}
          </>
        )
      }
    >
      {error && (
        <p className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-600 dark:bg-rose-500/15 dark:text-rose-400">
          {error}
        </p>
      )}

      {view === "list" ? (
        <>
          {!loading && docs.length > 0 && (
            <div className="flex flex-wrap items-center gap-2">
              <input
                className={`${inputCls} w-full sm:w-64`}
                value={keyword}
                onChange={(e) => setKeyword(e.target.value)}
                placeholder={t("docs.searchPlaceholder")}
              />
              {keyword && (
                <button
                  type="button"
                  onClick={() => setKeyword("")}
                  className="whitespace-nowrap rounded-lg px-3 py-2 text-sm text-slate-500 transition hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
                >
                  {t("common.clear")}
                </button>
              )}
              <span className="text-xs text-slate-500 dark:text-slate-400">
                {t("docs.count", { visible: visibleDocs.length, total: docs.length })}
              </span>
            </div>
          )}
          {loading ? (
            <p className="py-8 text-center text-sm text-slate-500 dark:text-slate-400">
              {t("common.loading")}
            </p>
          ) : docs.length === 0 ? (
            <div className="rounded-xl border border-dashed border-slate-300 px-6 py-10 text-center text-sm text-slate-500 dark:border-slate-700 dark:text-slate-400">
              {t("docs.empty")}
              {canCreate ? t("docs.emptyHint") : ""}
            </div>
          ) : visibleDocs.length === 0 ? (
            <p className="py-8 text-center text-sm text-slate-500 dark:text-slate-400">
              {t("docs.noMatch", { keyword: keyword.trim() })}
            </p>
          ) : (
            <ul className="divide-y divide-slate-100 overflow-hidden rounded-xl border border-slate-200 dark:divide-slate-800 dark:border-slate-700">
              {visibleDocs.map((d) => (
                <li key={d.id}>
                  <button
                    type="button"
                    onClick={() => openDoc(d)}
                    className="block w-full px-4 py-3 text-left transition hover:bg-slate-50 dark:hover:bg-slate-800/60"
                  >
                    <p className="truncate text-sm font-medium text-slate-800 dark:text-slate-100">
                      {d.title}
                    </p>
                    <p className="mt-1 truncate text-xs text-slate-500 dark:text-slate-400">
                      {richTextToText(d.content) || t("docs.noContent")}
                    </p>
                    <p className="mt-1 flex flex-wrap items-center gap-x-3 text-xs text-slate-500 dark:text-slate-400">
                      <span>{d.creator?.username ?? t("common.unknown")}</span>
                      <span>{t("docs.updatedAt", { time: formatDateTime(d.updatedAt, locale) })}</span>
                      {!!d.attachments?.length && <span>📎 {d.attachments.length}</span>}
                    </p>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </>
      ) : (
        <>
          {formError && (
            <p className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-600 dark:bg-rose-500/15 dark:text-rose-400">
              {formError}
            </p>
          )}

          {editing ? (
            <>
              <Field label={t("docs.field.title")}>
                <input
                  className={inputCls}
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder={t("docs.field.titlePlaceholder")}
                  autoFocus={!current}
                />
              </Field>
              <Field label={t("docs.field.content")}>
                <RichTextEditor
                  value={content}
                  onChange={setContent}
                  placeholder={t("docs.field.contentPlaceholder")}
                />
              </Field>
              <Field label={t("docs.field.attachments")}>
                <AttachmentList items={attachments} onChange={setAttachments} />
              </Field>
              {current && (
                <Field label={t("docs.field.creator")}>
                  <p className={valueCls}>{current.creator?.username ?? t("common.unknown")}</p>
                </Field>
              )}
            </>
          ) : (
            current && (
              <>
                <Field label={t("docs.field.title")}>
                  <p className={`${valueCls} font-medium text-slate-800 dark:text-slate-100`}>
                    {title}
                  </p>
                </Field>
                <Field label={t("docs.field.content")}>
                  <RichTextView html={content} empty={t("docs.noContent")} />
                </Field>
                <Field label={t("docs.field.attachments")}>
                  <AttachmentList items={attachments} onChange={setAttachments} readOnly />
                </Field>
                <div className="grid grid-cols-2 gap-3">
                  <Field label={t("docs.field.creator")}>
                    <p className={valueCls}>{current.creator?.username ?? t("common.unknown")}</p>
                  </Field>
                  <Field label={t("docs.field.updatedAt")}>
                    <p className={valueCls}>{formatDateTime(current.updatedAt, locale)}</p>
                  </Field>
                </div>
              </>
            )
          )}
        </>
      )}
    </Modal>
  );
}
