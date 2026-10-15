import { getCurrentLang, tNow } from "@/lib/i18n";
import type {
  Attachment,
  Department,
  LogPage,
  OperationLog,
  Project,
  ProjectDoc,
  Task,
  TaskStatus,
  TrashPayload,
  User,
} from "./types";

export const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8281/api";

const TOKEN_KEY = "mini-kanban_token";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null) {
  if (typeof window === "undefined") return;
  if (token) window.localStorage.setItem(TOKEN_KEY, token);
  else window.localStorage.removeItem(TOKEN_KEY);
}

// ---- 附件下载票据 ----
// 下载路由需要登录态，而 <img src> / <a href> 由浏览器直接发起、带不了 Authorization 头，
// 凭据只能放进 URL；为不让登录 token 出现在 URL 里（泄露即账号失守），
// 这里换一张「只能读附件」的短时效票据拼进下载地址，由后端签发（POST /files/ticket）。
let fileTicket = "";
let fileTicketExp = 0; // 票据到期时间（毫秒时间戳）
let fileTicketReq: Promise<string> | null = null;

/** ensureFileTicket 取回（并缓存）下载票据；并发调用共用同一个请求 */
export async function ensureFileTicket(): Promise<string> {
  // 剩余不足 5 分钟就换一张，避免页面开着时图片 / 附件突然失效
  if (fileTicket && Date.now() < fileTicketExp - 5 * 60 * 1000) return fileTicket;
  if (fileTicketReq) return fileTicketReq;
  const token = getToken();
  if (!token) return "";
  fileTicketReq = (async () => {
    try {
      const res = await fetch(`${API_BASE}/files/ticket`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
      });
      if (!res.ok) return "";
      const data = (await res.json()) as { ticket: string; expiresIn: number };
      fileTicket = data.ticket;
      fileTicketExp = Date.now() + data.expiresIn * 1000;
      return fileTicket;
    } catch {
      return "";
    } finally {
      fileTicketReq = null;
    }
  })();
  return fileTicketReq;
}

/** clearFileTicket 登出 / 切换账号时丢弃缓存的票据 */
export function clearFileTicket() {
  fileTicket = "";
  fileTicketExp = 0;
}

/** fileUrl 把后端返回的相对附件地址（/api/files/xxx）补全为可直接访问的地址，
 *  并带上当前的下载票据（票据未取回时先给出地址，取回后重渲染会补上） */
export function fileUrl(path: string): string {
  if (!path) return "";
  if (/^https?:\/\//i.test(path)) return path;
  const url = `${API_BASE.replace(/\/api\/?$/, "")}${path}`;
  if (!fileTicket) return url;
  return `${url}${url.includes("?") ? "&" : "?"}ticket=${encodeURIComponent(fileTicket)}`;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    // 后端的错误信息 / 提示按这个头返回对应语言（服务端另有 cookie 兜底，
    // 但附件地址这类由浏览器直接发起的请求带不上自定义头，只能靠 cookie）
    "Accept-Language": getCurrentLang(),
  };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: { ...headers, ...(options.headers as object) },
    cache: "no-store",
  });

  const text = await res.text();
  const data = text ? JSON.parse(text) : null;

  if (res.status === 401) {
    setToken(null);
    if (typeof window !== "undefined" && !window.location.pathname.startsWith("/login")) {
      window.location.href = "/login";
    }
  }
  if (!res.ok) {
    throw new Error(data?.error || tNow("common.requestFailed", { status: res.status }));
  }
  return data as T;
}

const body = (data: unknown) => JSON.stringify(data);

// 路径参数里的 id 一律按字符串拼接：users 是雪花 ID、其他表是 UUID，
// 两者都不能经过 Number()（雪花 ID 超出安全整数、UUID 会变成 NaN）。
export const api = {
  login: (email: string, password: string) =>
    request<{ token: string; user: User }>("/auth/login", {
      method: "POST",
      body: body({ email, password }),
    }),

  logout: () => request<{ message: string }>("/auth/logout", { method: "POST" }),

  me: () => request<User>("/auth/me"),

  // 部门
  listDepartments: () => request<Department[]>("/departments"),
  getDepartment: (id: string) => request<Department>(`/departments/${id}`),
  createDepartment: (data: { name: string; remark?: string }) =>
    request<Department>("/departments", { method: "POST", body: body(data) }),
  updateDepartment: (id: string, data: { name: string; remark?: string }) =>
    request<Department>(`/departments/${id}`, { method: "PUT", body: body(data) }),
  deleteDepartment: (id: string) =>
    request<{ message: string }>(`/departments/${id}`, { method: "DELETE" }),

  // 用户
  listUsers: () => request<User[]>("/users"),
  createUser: (data: Partial<User> & { password: string }) =>
    request<User>("/users", { method: "POST", body: body(data) }),
  updateUser: (id: string, data: Partial<User> & { password?: string }) =>
    request<User>(`/users/${id}`, { method: "PUT", body: body(data) }),
  deleteUser: (id: string) => request<{ message: string }>(`/users/${id}`, { method: "DELETE" }),

  // 项目
  listProjects: () => request<Project[]>("/projects"),
  getProject: (id: string) => request<Project>(`/projects/${id}`),
  createProject: (data: { title: string; description?: string; memberIds?: string[] }) =>
    request<Project>("/projects", { method: "POST", body: body(data) }),
  updateProject: (id: string, data: { title?: string; description?: string; memberIds?: string[] }) =>
    request<Project>(`/projects/${id}`, { method: "PUT", body: body(data) }),
  deleteProject: (id: string) =>
    request<{ message: string }>(`/projects/${id}`, { method: "DELETE" }),

  // 项目资料（需求文档 / 会议纪要 / 设计稿 / 项目密钥 / 测试账密等：标题 + 富文本正文 + 附件）
  listProjectDocs: (projectId: string) => request<ProjectDoc[]>(`/projects/${projectId}/docs`),
  createProjectDoc: (
    projectId: string,
    data: { title: string; content?: string; attachments?: Attachment[] },
  ) => request<ProjectDoc>(`/projects/${projectId}/docs`, { method: "POST", body: body(data) }),
  updateProjectDoc: (
    id: string,
    data: { title?: string; content?: string; attachments?: Attachment[] },
  ) => request<ProjectDoc>(`/docs/${id}`, { method: "PUT", body: body(data) }),
  deleteProjectDoc: (id: string) =>
    request<{ message: string }>(`/docs/${id}`, { method: "DELETE" }),

  // 任务
  listTasks: (projectId: string) => request<Task[]>(`/projects/${projectId}/tasks`),
  createTask: (
    projectId: string,
    data: {
      title: string;
      description?: string;
      assigneeId?: string | null;
      dueDate?: string | null;
      status?: TaskStatus;
      attachments?: Attachment[];
    },
  ) => request<Task>(`/projects/${projectId}/tasks`, { method: "POST", body: body(data) }),
  updateTask: (
    id: string,
    data: {
      title?: string;
      description?: string;
      assigneeId?: string | null;
      dueDate?: string | null;
      status?: TaskStatus;
      attachments?: Attachment[];
    },
  ) => request<Task>(`/tasks/${id}`, { method: "PUT", body: body(data) }),
  deleteTask: (id: string) => request<{ message: string }>(`/tasks/${id}`, { method: "DELETE" }),
  moveTask: (id: string, status: TaskStatus, index: number) =>
    request<Task>(`/tasks/${id}/move`, { method: "PATCH", body: body({ status, index }) }),

  // 回收箱：软删除数据的集中管理
  listTrash: () => request<TrashPayload>("/trash"),
  restoreProject: (id: string) =>
    request<{ message: string }>(`/trash/projects/${id}/restore`, { method: "POST" }),
  purgeProject: (id: string) =>
    request<{ message: string }>(`/trash/projects/${id}`, { method: "DELETE" }),
  restoreTask: (id: string) =>
    request<{ message: string }>(`/trash/tasks/${id}/restore`, { method: "POST" }),
  purgeTask: (id: string) =>
    request<{ message: string }>(`/trash/tasks/${id}`, { method: "DELETE" }),
  restoreDoc: (id: string) =>
    request<{ message: string }>(`/trash/docs/${id}/restore`, { method: "POST" }),
  purgeDoc: (id: string) =>
    request<{ message: string }>(`/trash/docs/${id}`, { method: "DELETE" }),

  // 操作日志（管理员）：分页 + 按对象类型 / 动作 / 关键词筛选。
  // 列表不带 snapshot（可能很大），需要看快照时再调 getLog 取单条。
  listLogs: (
    params: {
      targetType?: string;
      action?: string;
      keyword?: string;
      page?: number;
      pageSize?: number;
    } = {},
  ) => {
    const qs = new URLSearchParams();
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== null && v !== "") qs.set(k, String(v));
    });
    const query = qs.toString();
    return request<LogPage>(`/logs${query ? `?${query}` : ""}`);
  },
  getLog: (id: string) => request<OperationLog>(`/logs/${id}`),

  // 附件上传：multipart/form-data，一次可传多个文件（字段名 files）
  // 注意不能走 request()：那里会强制 JSON 内容类型，破坏 multipart 边界
  uploadFiles: async (files: File[]): Promise<Attachment[]> => {
    const fd = new FormData();
    files.forEach((f) => fd.append("files", f));
    const token = getToken();
    const res = await fetch(`${API_BASE}/uploads`, {
      method: "POST",
      body: fd,
      headers: {
        "Accept-Language": getCurrentLang(),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      cache: "no-store",
    });
    const text = await res.text();
    const data = text ? JSON.parse(text) : null;
    if (res.status === 401) {
      setToken(null);
      if (typeof window !== "undefined" && !window.location.pathname.startsWith("/login")) {
        window.location.href = "/login";
      }
    }
    if (!res.ok) {
      throw new Error(
        data?.error || tNow("common.uploadRequestFailed", { status: res.status }),
      );
    }
    return data as Attachment[];
  },
};
