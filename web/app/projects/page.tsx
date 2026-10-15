"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { AuthGuard } from "@/components/AuthGuard";
import { useAuth } from "@/components/AuthProvider";
import { useConfirm } from "@/components/ConfirmProvider";
import { useI18n } from "@/components/I18nProvider";
import { ProjectModal, type ProjectFormValue } from "@/components/ProjectModal";
import { api } from "@/lib/api";
import { tNow } from "@/lib/i18n";
import type { Department, Project, User } from "@/lib/types";

export default function ProjectsPage() {
  const { user } = useAuth();
  const { t } = useI18n();
  const confirm = useConfirm();
  const [projects, setProjects] = useState<Project[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [error, setError] = useState("");
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Project | null>(null);

  const load = useCallback(async () => {
    try {
      const [list, u, d] = await Promise.all([
        api.listProjects(),
        api.listUsers(),
        api.listDepartments(),
      ]);
      setProjects(list);
      setUsers(u);
      setDepartments(d);
    } catch (err) {
      setError(err instanceof Error ? err.message : tNow("common.loadFailed"));
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  // 项目编辑 / 删除权限：管理员或项目创建者（后端同样会校验）
  const canManage = (p: Project) => user?.role === "admin" || p.ownerId === user?.id;

  const openCreate = () => {
    setEditing(null);
    setOpen(true);
  };

  const openEdit = (p: Project) => {
    setEditing(p);
    setOpen(true);
  };

  // 表单状态与校验都在 ProjectModal 内，这里只负责调接口与刷新
  const save = async (v: ProjectFormValue) => {
    try {
      if (editing) await api.updateProject(editing.id, v);
      else await api.createProject(v);
      setOpen(false);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("common.saveFailed"));
    }
  };

  const remove = async (p: Project) => {
    const ok = await confirm({
      title: t("project.deleteTitle"),
      description: t("project.deleteDesc", { title: p.title }),
      note: t("project.deleteNote"),
      confirmText: t("common.delete"),
      danger: true,
    });
    if (!ok) return;
    try {
      await api.deleteProject(p.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("common.deleteFailed"));
    }
  };

  return (
    <AuthGuard>
      <main className="mx-auto max-w-[1400px] px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">
              {t("projects.title")}
            </h1>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
              {t("projects.subtitle")}
            </p>
          </div>
          <button
            onClick={openCreate}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700"
          >
            {`+ ${t("projects.create")}`}
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-rose-50 px-4 py-2 text-sm text-rose-600 dark:bg-rose-500/15 dark:text-rose-400">
            {error}
          </div>
        )}

        {projects.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-12 text-center text-slate-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-400">
            {t("projects.empty")}
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {projects.map((p) => (
              <div
                key={p.id}
                className="group rounded-2xl border border-slate-200 bg-white p-5 shadow-card transition hover:border-indigo-300 dark:border-slate-800 dark:bg-slate-900 dark:hover:border-indigo-500/60"
              >
                <Link href={`/projects/${p.id}`} className="block">
                  <h2 className="text-base font-semibold text-slate-900 group-hover:text-indigo-600 dark:text-slate-100 dark:group-hover:text-indigo-400">
                    {p.title}
                  </h2>
                  <p className="mt-2 line-clamp-3 min-h-[3rem] text-sm text-slate-500 dark:text-slate-400">
                    {p.description || t("projects.noDescription")}
                  </p>
                </Link>
                <div className="mt-4 flex items-center justify-between text-xs text-slate-400 dark:text-slate-500">
                  <span>
                    {t("projects.summary", {
                      tasks: p.taskCount ?? 0,
                      members: p.members?.length ?? 0,
                    })}
                  </span>
                  <div className="flex gap-2">
                    {canManage(p) && (
                      <>
                        <button
                          onClick={() => openEdit(p)}
                          className="rounded px-2 py-1 text-slate-500 transition hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
                        >
                          {t("common.edit")}
                        </button>
                        <button
                          onClick={() => remove(p)}
                          className="rounded px-2 py-1 text-rose-500 transition hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-500/15"
                        >
                          {t("common.delete")}
                        </button>
                      </>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {open && (
          <ProjectModal
            key={editing ? `project-${editing.id}` : "project-new"}
            open={open}
            project={editing}
            users={users}
            departments={departments}
            onClose={() => setOpen(false)}
            onSubmit={save}
          />
        )}
      </main>
    </AuthGuard>
  );
}
