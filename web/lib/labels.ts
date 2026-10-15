import type { MessageKey, TFunction } from "@/lib/i18n";
import type { LogAction, LogTargetType, Role, TaskStatus } from "@/lib/types";

/**
 * 业务枚举 → 展示文案。
 *
 * 枚举值本身（todo / admin / project...）是接口契约，后端只下发这些稳定值，
 * 中文 / 英文只是展示层的事，所以映射集中在文案层，接口与数据库都不动。
 * 遇到未知值一律原样回显，避免新旧版本不一致时界面出现空白。
 */

const STATUS_KEYS: Record<TaskStatus, MessageKey> = {
  todo: "column.todo",
  doing: "column.doing",
  done: "column.done",
  accepted: "column.accepted",
};

const ROLE_KEYS: Record<Role, MessageKey> = {
  admin: "role.admin",
  member: "role.member",
};

const LOG_TARGET_KEYS: Record<LogTargetType, MessageKey> = {
  project: "logTarget.project",
  task: "logTarget.task",
  doc: "logTarget.doc",
  user: "logTarget.user",
  department: "logTarget.department",
};

const LOG_ACTION_KEYS: Record<LogAction, MessageKey> = {
  create: "logAction.create",
  update: "logAction.update",
  move: "logAction.move",
  delete: "logAction.delete",
  restore: "logAction.restore",
  purge: "logAction.purge",
};

/** 看板列名 / 任务状态 */
export function statusLabel(t: TFunction, status: string): string {
  const key = STATUS_KEYS[status as TaskStatus];
  return key ? t(key) : status;
}

/** 角色名 */
export function roleLabel(t: TFunction, role: string): string {
  const key = ROLE_KEYS[role as Role];
  return key ? t(key) : role;
}

/** 操作日志的对象类型 */
export function logTargetLabel(t: TFunction, targetType: string): string {
  const key = LOG_TARGET_KEYS[targetType as LogTargetType];
  return key ? t(key) : targetType;
}

/** 操作日志的动作 */
export function logActionLabel(t: TFunction, action: string): string {
  const key = LOG_ACTION_KEYS[action as LogAction];
  return key ? t(key) : action;
}

/** 日志筛选下拉的选项顺序（对象类型 / 动作） */
export const LOG_TARGET_TYPES: LogTargetType[] = [
  "project",
  "task",
  "doc",
  "user",
  "department",
];

export const LOG_ACTIONS: LogAction[] = [
  "create",
  "update",
  "move",
  "delete",
  "restore",
  "purge",
];
