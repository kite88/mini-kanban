"use client";

import { useState } from "react";
import { useAuth } from "@/components/AuthProvider";
import { useI18n } from "@/components/I18nProvider";
import { Field, Modal, inputCls } from "@/components/Modal";
import { MemberPicker } from "@/components/MemberPicker";
import type { Department, Project, User } from "@/lib/types";

export interface ProjectFormValue {
  title: string;
  description: string;
  memberIds: string[];
}

// ProjectModal 项目表单弹窗：项目列表页与项目看板页共用（与任务弹窗同一套样式与交互）
export function ProjectModal({
  open,
  project,
  users,
  departments = [],
  onClose,
  onSubmit,
}: {
  open: boolean;
  project: Project | null;
  users: User[];
  departments?: Department[];
  onClose: () => void;
  onSubmit: (v: ProjectFormValue) => Promise<void> | void;
}) {
  const { user } = useAuth();
  const { t } = useI18n();
  const [title, setTitle] = useState(project?.title ?? "");
  const [description, setDescription] = useState(project?.description ?? "");
  const [memberIds, setMemberIds] = useState<string[]>(
    project ? (project.members ?? []).map((m) => m.id) : user ? [user.id] : [],
  );
  const [saving, setSaving] = useState(false);

  // 创建人固定关联、不可取消（新建时是当前用户，编辑时是项目创建者）
  const lockedIds = project ? [project.ownerId] : user ? [user.id] : [];

  const save = async () => {
    if (!title.trim() || saving) return;
    setSaving(true);
    try {
      await onSubmit({
        title,
        description,
        memberIds: Array.from(new Set([...lockedIds, ...memberIds])),
      });
    } catch {
      /* 失败提示由调用方的 error 状态负责 */
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal
      open={open}
      title={project ? t("project.editTitle") : t("project.createTitle")}
      onClose={onClose}
      footer={
        <>
          <button
            onClick={onClose}
            className="rounded-lg border border-slate-200 px-4 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {t("common.cancel")}
          </button>
          <button
            onClick={save}
            disabled={saving || !title.trim()}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700 disabled:opacity-60"
          >
            {saving ? t("common.saving") : t("common.save")}
          </button>
        </>
      }
    >
      <Field label={t("project.field.title")}>
        <input
          className={inputCls}
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder={t("project.field.titlePlaceholder")}
          autoFocus
        />
      </Field>
      <Field label={t("project.field.description")}>
        <textarea
          className={inputCls}
          rows={4}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder={t("project.field.descriptionPlaceholder")}
        />
      </Field>
      <Field label={t("project.field.members")}>
        <MemberPicker
          users={users}
          departments={departments}
          value={memberIds}
          lockedIds={lockedIds}
          onChange={setMemberIds}
        />
      </Field>
    </Modal>
  );
}
