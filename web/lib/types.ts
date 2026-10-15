export type Role = "admin" | "member";

// ID 约定（与后端 models 保持一致）：
//   - users 表主键是 64 位雪花 ID，超出 JS 安全整数，接口按字符串收发；
//   - 其他表（departments / projects / tasks / project_docs / 附件 / 操作日志）是 UUID v7 字符串。
// 因此所有 id 与外键列在类型上一律是 string，前端不做任何 Number() 转换。

export interface Department {
  id: string;
  name: string;
  remark: string;
  userCount?: number;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: string;
  email: string;
  username: string;
  /** 所属部门 id（UUID）；null 表示未分配 */
  departmentId: string | null;
  role: Role;
  createdAt: string;
}

export interface Project {
  id: string;
  title: string;
  description: string;
  ownerId: string;
  owner?: User;
  members?: User[];
  taskCount?: number;
  createdAt: string;
  /** 软删除时间：为空表示未被删除（回收箱里的条目一定非空） */
  deletedAt?: string | null;
}

export type TaskStatus = "todo" | "doing" | "done" | "accepted";

/** 任务附件：url 是相对地址（/api/files/xxx），用 fileUrl() 补全后再访问 */
export interface Attachment {
  id?: string;
  name: string;
  url: string;
  size: number;
}

export interface Task {
  id: string;
  projectId: string;
  title: string;
  /** 富文本 HTML（服务端已净化，内含的图片用 /api/files/xxx 相对地址） */
  description: string;
  assigneeId: string | null;
  assignee?: User | null;
  creatorId: string;
  creator?: User | null;
  status: TaskStatus;
  position: number;
  dueDate: string | null;
  attachments?: Attachment[];
  createdAt: string;
  /** 软删除时间：为空表示未被删除（回收箱里的条目一定非空） */
  deletedAt?: string | null;
}

/** 项目资料：项目下的文档型资料（需求文档 / 会议纪要 / 设计稿 / 项目密钥 / 测试账密等），标题 + 富文本正文 + 附件 */
export interface ProjectDoc {
  id: string;
  projectId: string;
  title: string;
  /** 富文本 HTML（服务端已净化，内含的图片用 /api/files/xxx 相对地址） */
  content: string;
  creatorId: string;
  creator?: User | null;
  attachments?: Attachment[];
  createdAt: string;
  updatedAt: string;
  /** 软删除时间：为空表示未被删除 */
  deletedAt?: string | null;
}

// 看板四列：接口只认 id（todo/doing/done/accepted，与后端 models 一致），
// 展示用的列名 / 角色名等文案在 lib/labels.ts + lib/i18n 里按语言取。
export const COLUMNS: { id: TaskStatus; accent: string }[] = [
  { id: "todo", accent: "bg-slate-400" },
  { id: "doing", accent: "bg-blue-500" },
  { id: "done", accent: "bg-emerald-500" },
  { id: "accepted", accent: "bg-violet-500" },
];

/** 回收箱里的项目条目：taskCount / docCount 是随它一起被删、恢复时会一并回来的数量 */
export interface TrashProject extends Project {
  taskCount: number;
  docCount: number;
}

/** 回收箱里的任务条目：任务表里只有 projectId，标题由后端补上便于展示 */
export interface TrashTask extends Task {
  projectTitle: string;
}

/** 回收箱里的项目资料条目：资料表里只有 projectId，标题由后端补上便于展示 */
export interface TrashDoc extends ProjectDoc {
  projectTitle: string;
}

export interface TrashPayload {
  projects: TrashProject[];
  tasks: TrashTask[];
  docs: TrashDoc[];
}

/** 操作日志的对象类型（与后端 models.Target* 一一对应） */
export type LogTargetType = "project" | "task" | "doc" | "user" | "department";

/** 操作日志的动作（与后端 models.Action* 一一对应） */
export type LogAction = "create" | "update" | "move" | "delete" | "restore" | "purge";

export interface OperationLog {
  id: string;
  /** 操作人 id：users 主键，雪花 ID 字符串；用户被物理删除后由名称快照兜底 */
  operatorId: string;
  /** 操作人用户名快照：操作人被删除后日志依旧可读 */
  operatorName: string;
  targetType: LogTargetType;
  /** 目标 id：项目 / 任务 / 资料是 UUID 字符串，用户 / 部门是雪花 ID 字符串 */
  targetId: string;
  targetName: string;
  action: LogAction;
  summary: string;
  /**
   * 数据快照：只有物理删除（删用户 / 删部门 / 回收箱彻底删除）才有。
   * 列表接口刻意不下发，由详情接口（getLog）单独返回。
   */
  snapshot?: unknown;
  createdAt: string;
}

export interface LogPage {
  items: OperationLog[];
  total: number;
  page: number;
  pageSize: number;
}

// 对象类型 / 动作的展示文案在 lib/labels.ts（logTargetLabel / logActionLabel），
// 日志摘要（summary）与变更项是后端写入时落库的文本，保持原样展示。
