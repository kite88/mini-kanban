"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, usePathname, useRouter } from "next/navigation";
import { AuthGuard } from "@/components/AuthGuard";
import { useAuth } from "@/components/AuthProvider";
import { Board } from "@/components/Board";
import { useConfirm } from "@/components/ConfirmProvider";
import { useI18n } from "@/components/I18nProvider";
import { ProjectDocsModal } from "@/components/ProjectDocsModal";
import { ProjectModal, type ProjectFormValue } from "@/components/ProjectModal";
import { TaskModal, type TaskFormValue } from "@/components/TaskModal";
import { api } from "@/lib/api";
import { tNow, type MessageKey } from "@/lib/i18n";
import type { Department, Project, Task, TaskStatus, User } from "@/lib/types";

type TaskFilter = "all" | "mine" | "assigned";

const TASK_FILTERS: { id: TaskFilter; labelKey: MessageKey }[] = [
  { id: "all", labelKey: "board.filter.all" },
  { id: "mine", labelKey: "board.filter.mine" },
  { id: "assigned", labelKey: "board.filter.assigned" },
];

/**
 * 取当前看板的项目 id。
 *
 * 优先用地址栏路径、useParams() 兜底的原因：静态导出（output: "export"）下看板路由
 * 只有一份预渲染的占位页 /projects/_/，真实 id 的深链 /projects/<uuid>/ 由静态服务器
 * 映射到它（见 server/webstatic.go 的 boardPlaceholderFile）。此时 useParams() 读的是
 * 构建期写死的路由树（值是占位 id），只有 usePathname() 跟着浏览器地址走。
 * 常规的服务端渲染与站内跳转下两者一致，取路径同样正确。
 */
function resolveProjectId(fromParams: string | undefined, pathname: string | null): string {
  const segments = (pathname ?? "").split("/").filter(Boolean);
  // /projects/<id> 与 /projects/<id>/ 都取第二段；id 是 UUID，不含斜杠
  if (segments[0] === "projects" && segments[1]) return decodeURIComponent(segments[1]);
  return fromParams ?? "";
}

export default function ProjectBoardPage() {
  const params = useParams<{ id: string }>();
  const pathname = usePathname();
  const router = useRouter();
  const { user } = useAuth();
  const { t } = useI18n();
  const confirm = useConfirm();
  // 项目 id 是 UUID 字符串：不能走 Number()（会变成 NaN），保持字符串直接透传给接口
  const projectId = useMemo(() => resolveProjectId(params?.id, pathname), [params?.id, pathname]);

  const [project, setProject] = useState<Project | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [error, setError] = useState("");
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [projectModalOpen, setProjectModalOpen] = useState(false);
  const [docsModalOpen, setDocsModalOpen] = useState(false);
  const [defaultStatus, setDefaultStatus] = useState<TaskStatus>("todo");
  const [filter, setFilter] = useState<TaskFilter>("all");

  // 筛选只看当前用户的维度。用 useMemo 保持引用稳定：
  // Board 里的 columns 依赖 tasks 的引用，每次渲染都新建数组会导致它反复重置。
  const visibleTasks = useMemo(() => {
    if (!user) return tasks;
    if (filter === "mine") return tasks.filter((t) => t.creatorId === user.id);
    if (filter === "assigned") return tasks.filter((t) => t.assigneeId === user.id);
    return tasks;
  }, [tasks, filter, user]);

  const load = useCallback(async () => {
    try {
      const [p, t, u, d] = await Promise.all([
        api.getProject(projectId),
        api.listTasks(projectId),
        api.listUsers(),
        // 部门用于在「关联人」下拉里标注所属部门；取不到时降级为空，不影响看板
        api.listDepartments().catch(() => [] as Department[]),
      ]);
      setProject(p);
      setTasks(t);
      setUsers(u);
      setDepartments(d);
    } catch (err) {
      setError(err instanceof Error ? err.message : tNow("common.loadFailed"));
    }
  }, [projectId]);

  useEffect(() => {
    if (projectId) load();
  }, [projectId, load]);

  const openCreate = useCallback((status: TaskStatus) => {
    setEditingTask(null);
    setDefaultStatus(status);
    setModalOpen(true);
  }, []);

  const openEdit = useCallback((task: Task) => {
    setEditingTask(task);
    setModalOpen(true);
  }, []);

  const submit = async (v: TaskFormValue) => {
    if (editingTask) {
      await api.updateTask(editingTask.id, v);
    } else {
      await api.createTask(projectId, { ...v, status: v.status || defaultStatus });
    }
    setModalOpen(false);
    await load();
  };

  // 编辑项目：与项目列表页共用同一个弹窗（标题 / 描述 / 关联成员）
  const submitProject = async (v: ProjectFormValue) => {
    try {
      await api.updateProject(projectId, v);
      setProjectModalOpen(false);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("common.saveFailed"));
    }
  };

  // 任务操作权限：管理员、任务创建人、被指派人、项目创建人
  // （后端 canManageTask 是同一套规则，查看/编辑/拖拽/删除都按它判断）
  const isProjectOwner = !!project && !!user && project.ownerId === user.id;

  const canEdit = useCallback(
    (t: Task | null) =>
      !!t &&
      (user?.role === "admin" ||
        t.creatorId === user?.id ||
        t.assigneeId === user?.id ||
        isProjectOwner),
    [user, isProjectOwner],
  );

  // 删除与编辑同规则：四类角色都能删（后端 DeleteTask 同样会校验）
  const canDelete = canEdit;

  // 建任务权限：管理员或项目成员（被指派到该项目但非成员的参与人不能新建任务）
  const canCreateTask =
    user?.role === "admin" || (!!user && (project?.members ?? []).some((m) => m.id === user.id));

  // 项目编辑 / 删除权限：管理员或项目创建者（后端同样会校验）
  const canManageProject = !!project && (user?.role === "admin" || project.ownerId === user?.id);

  const remove = async () => {
    if (!editingTask) return;
    const ok = await confirm({
      title: t("board.deleteTaskTitle"),
      description: t("board.deleteTaskDesc", { title: editingTask.title }),
      note: t("project.deleteNote"),
      confirmText: t("common.delete"),
      danger: true,
    });
    if (!ok) return;
    await api.deleteTask(editingTask.id);
    setModalOpen(false);
    await load();
  };

  return (
    <AuthGuard>
      <main className="mx-auto max-w-[1400px] px-6 py-6">
        <div className="mb-5 flex items-start justify-between gap-4">
          <div>
            <Link
              href="/projects"
              className="text-sm text-slate-500 transition hover:text-indigo-600 dark:text-slate-400 dark:hover:text-indigo-400"
            >
              {t("board.backToProjects")}
            </Link>
            <h1 className="mt-1 text-2xl font-semibold text-slate-900 dark:text-slate-100">
              {project?.title ?? t("common.loading")}
            </h1>
            {project?.description && (
              <p className="mt-1 max-w-2xl text-sm text-slate-500 dark:text-slate-400">
                {project.description}
              </p>
            )}
            {!!project?.members?.length && (
              <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">
                {t("board.members", { count: project.members.length })}
                {project.members.map((m) => m.username).join(t("common.listSeparator"))}
              </p>
            )}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setDocsModalOpen(true)}
              className="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {t("board.docs")}
            </button>
            {canManageProject && (
              <>
                <button
                  onClick={() => setProjectModalOpen(true)}
                  className="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800"
                >
                  {t("board.editProject")}
                </button>
                <button
                  onClick={async () => {
                    const ok = await confirm({
                      title: t("project.deleteTitle"),
                      description: t("project.deleteDesc", { title: project?.title ?? "" }),
                      note: t("project.deleteNote"),
                      confirmText: t("common.delete"),
                      danger: true,
                    });
                    if (!ok) return;
                    try {
                      await api.deleteProject(projectId);
                      router.push("/projects");
                    } catch (e) {
                      setError(e instanceof Error ? e.message : t("common.deleteFailed"));
                    }
                  }}
                  className="rounded-lg border border-rose-200 bg-white px-4 py-2 text-sm text-rose-600 transition hover:bg-rose-50 dark:border-rose-500/40 dark:bg-slate-900 dark:text-rose-400 dark:hover:bg-rose-500/15"
                >
                  {t("board.deleteProject")}
                </button>
              </>
            )}
            <button
              onClick={() => openCreate("todo")}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700"
            >
              {t("board.createTask")}
            </button>
          </div>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-rose-50 px-4 py-2 text-sm text-rose-600 dark:bg-rose-500/15 dark:text-rose-400">
            {error}
          </div>
        )}

        <div className="mb-4 flex flex-wrap items-center gap-2">
          {TASK_FILTERS.map((f) => (
            <button
              key={f.id}
              onClick={() => setFilter(f.id)}
              className={`rounded-lg px-3 py-1 text-xs transition ${
                filter === f.id
                  ? "bg-indigo-600 text-white"
                  : "border border-slate-200 text-slate-600 hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
              }`}
            >
              {t(f.labelKey)}
            </button>
          ))}
          <span className="ml-1 text-xs text-slate-400 dark:text-slate-500">
            {t("board.taskSummary", { visible: visibleTasks.length, total: tasks.length })}
          </span>
        </div>

        {visibleTasks.length === 0 && (
          <div className="mb-4 rounded-lg bg-slate-100 px-4 py-2 text-xs text-slate-500 dark:bg-slate-800 dark:text-slate-400">
            {filter === "all" ? t("board.emptyAll") : t("board.emptyFiltered")}
          </div>
        )}

        <Board
          tasks={visibleTasks}
          onReload={load}
          onEdit={openEdit}
          onCreate={openCreate}
          onError={setError}
          canDrag={canEdit}
          canCreate={canCreateTask}
        />

        {modalOpen && (
          <TaskModal
            key={editingTask ? `task-${editingTask.id}` : `new-${defaultStatus}`}
            open={modalOpen}
            task={editingTask}
            users={users}
            departments={departments}
            defaultStatus={defaultStatus}
            onClose={() => setModalOpen(false)}
            onSubmit={submit}
            canEdit={editingTask ? canEdit(editingTask) : true}
            onDelete={canDelete(editingTask) ? remove : undefined}
          />
        )}

        {projectModalOpen && (
          <ProjectModal
            key={`project-${projectId}`}
            open={projectModalOpen}
            project={project}
            users={users}
            departments={departments}
            onClose={() => setProjectModalOpen(false)}
            onSubmit={submitProject}
          />
        )}

        {docsModalOpen && project && (
          <ProjectDocsModal
            key={`docs-${projectId}`}
            open={docsModalOpen}
            project={project}
            onClose={() => setDocsModalOpen(false)}
          />
        )}
      </main>
    </AuthGuard>
  );
}
