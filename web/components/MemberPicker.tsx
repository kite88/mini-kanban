"use client";

import { useMemo } from "react";
import { useI18n } from "@/components/I18nProvider";
import type { Department, User } from "@/lib/types";

/**
 * 成员选择器：按部门分组勾选，勾部门标题 = 批量选中该部门下所有成员。
 * 选择结果最终展开为具体用户 id 列表（部门本身不落库）。
 */
export function MemberPicker({
  users,
  departments,
  value,
  lockedIds = [],
  onChange,
}: {
  users: User[];
  departments: Department[];
  value: string[];
  lockedIds?: string[]; // 固定成员（如项目创建人），不可取消
  onChange: (ids: string[]) => void;
}) {
  const { t } = useI18n();
  const groups = useMemo(() => {
    // 部门 id 是 UUID、用户 id 是雪花 ID，都是字符串，Map 键按字符串比较
    const byDept = new Map<string | null, User[]>();
    for (const u of users) {
      const key = u.departmentId ?? null;
      const list = byDept.get(key);
      if (list) list.push(u);
      else byDept.set(key, [u]);
    }
    const nameOf = (id: string | null) =>
      id === null
        ? t("memberPicker.noDepartment")
        : (departments.find((d) => d.id === id)?.name ??
          t("memberPicker.departmentFallback", { id }));
    return Array.from(byDept.entries())
      // id 是字符串，排序用 localeCompare：减法排序在字符串上会得到 NaN
      .map(([id, list]) => ({
        id,
        name: nameOf(id),
        users: [...list].sort((a, b) => a.id.localeCompare(b.id)),
      }))
      .sort((a, b) =>
        a.id === null ? 1 : b.id === null ? -1 : a.id.localeCompare(b.id),
      );
  }, [users, departments, t]);

  // 固定成员（创建人）始终视为已选，避免外部数据没带回来时显示成未勾选
  const isChecked = (id: string) => value.includes(id) || lockedIds.includes(id);

  const toggle = (id: string) => {
    if (lockedIds.includes(id)) return;
    onChange(value.includes(id) ? value.filter((x) => x !== id) : [...value, id]);
  };

  const toggleGroup = (list: User[]) => {
    const ids = list.map((u) => u.id).filter((id) => !lockedIds.includes(id));
    const allSelected = ids.length > 0 && ids.every((id) => value.includes(id));
    onChange(
      allSelected
        ? value.filter((id) => !ids.includes(id))
        : Array.from(new Set([...value, ...ids])),
    );
  };

  return (
    <div>
      <div className="max-h-56 overflow-y-auto rounded-lg border border-slate-200 p-2 dark:border-slate-700">
        {groups.length === 0 && (
          <p className="px-1 py-2 text-xs text-slate-400 dark:text-slate-500">
            {t("memberPicker.empty")}
          </p>
        )}
        {groups.map((g) => {
          const ids = g.users.map((u) => u.id).filter((id) => !lockedIds.includes(id));
          const selected = ids.filter(isChecked).length;
          const all = ids.length > 0 && selected === ids.length;
          const some = selected > 0 && !all;
          return (
            <div key={String(g.id)} className="mb-1.5 last:mb-0">
              <label className="flex cursor-pointer items-center gap-2 rounded px-1 py-1 text-sm font-medium text-slate-700 hover:bg-slate-50 dark:text-slate-200 dark:hover:bg-slate-800">
                <input
                  type="checkbox"
                  className="h-4 w-4 accent-indigo-600"
                  checked={all}
                  ref={(el) => {
                    if (el) el.indeterminate = some;
                  }}
                  onChange={() => toggleGroup(g.users)}
                />
                {g.name}
                <span className="text-xs font-normal text-slate-400 dark:text-slate-500">
                  {selected}/{ids.length}
                </span>
              </label>
              <div className="ml-6 flex flex-wrap gap-x-4 gap-y-1 py-0.5">
                {g.users.map((u) => (
                  <label
                    key={u.id}
                    className="flex cursor-pointer items-center gap-1.5 text-xs text-slate-600 dark:text-slate-300"
                  >
                    <input
                      type="checkbox"
                      className="h-3.5 w-3.5 accent-indigo-600"
                      checked={isChecked(u.id)}
                      disabled={lockedIds.includes(u.id)}
                      onChange={() => toggle(u.id)}
                    />
                    {u.username}
                    {lockedIds.includes(u.id) && (
                      <span className="text-[10px] text-slate-400 dark:text-slate-500">
                        {t("memberPicker.creator")}
                      </span>
                    )}
                  </label>
                ))}
              </div>
            </div>
          );
        })}
      </div>
      <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">
        {t("memberPicker.selected", { count: new Set([...value, ...lockedIds]).size })}
      </p>
    </div>
  );
}
