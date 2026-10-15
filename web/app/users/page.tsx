"use client";

import { useCallback, useEffect, useState } from "react";
import { AuthGuard } from "@/components/AuthGuard";
import { useConfirm } from "@/components/ConfirmProvider";
import { useI18n } from "@/components/I18nProvider";
import { Field, Modal, inputCls } from "@/components/Modal";
import { api } from "@/lib/api";
import { formatDateTime } from "@/lib/date";
import { tNow } from "@/lib/i18n";
import { roleLabel } from "@/lib/labels";
import type { Department, Role, User } from "@/lib/types";
import { isValidEmail } from "@/lib/validate";

export default function UsersPage() {
  const { t, locale } = useI18n();
  const confirm = useConfirm();
  const [users, setUsers] = useState<User[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [error, setError] = useState("");
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [form, setForm] = useState({
    email: "",
    username: "",
    departmentId: "", // 空串表示未分配（部门 id 是 UUID，后端收到空串会写成 NULL）
    password: "",
    role: "member" as Role,
  });
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      const [list, depts] = await Promise.all([api.listUsers(), api.listDepartments()]);
      setUsers(list);
      setDepartments(depts);
    } catch (err) {
      setError(err instanceof Error ? err.message : tNow("common.loadFailed"));
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setEditing(null);
    setForm({ email: "", username: "", departmentId: "", password: "", role: "member" });
    setOpen(true);
  };

  const openEdit = (u: User) => {
    setEditing(u);
    setForm({
      email: u.email,
      username: u.username,
      departmentId: u.departmentId ?? "",
      password: "",
      role: u.role,
    });
    setOpen(true);
  };

  // 邮箱格式校验：输入时即时提示，保存前再拦一道（后端也会校验一次）
  const email = form.email.trim();
  const emailInvalid = form.email !== "" && !isValidEmail(form.email);

  const save = async () => {
    if (!isValidEmail(email)) {
      setError(t("validate.emailInvalid"));
      return;
    }
    setError("");
    setSaving(true);
    try {
      if (editing) {
        const payload: Partial<User> & { password?: string } = {
          email,
          username: form.username,
          departmentId: form.departmentId,
          role: form.role,
        };
        if (form.password) payload.password = form.password;
        await api.updateUser(editing.id, payload);
      } else {
        await api.createUser({ ...form, email });
      }
      setOpen(false);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("common.saveFailed"));
    } finally {
      setSaving(false);
    }
  };

  const remove = async (u: User) => {
    const ok = await confirm({
      title: t("users.deleteTitle"),
      description: t("users.deleteDesc", { name: u.username }),
      note: t("users.deleteNote"),
      confirmText: t("common.delete"),
      danger: true,
    });
    if (!ok) return;
    try {
      await api.deleteUser(u.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("common.deleteFailed"));
    }
  };

  return (
    <AuthGuard>
      <main className="mx-auto max-w-[1000px] px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">
              {t("users.title")}
            </h1>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{t("users.subtitle")}</p>
          </div>
          <button
            onClick={openCreate}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700"
          >
            {`+ ${t("users.create")}`}
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-rose-50 px-4 py-2 text-sm text-rose-600 dark:bg-rose-500/15 dark:text-rose-400">
            {error}
          </div>
        )}

        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-card dark:border-slate-800 dark:bg-slate-900">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-slate-500 dark:bg-slate-950/60 dark:text-slate-400">
              <tr>
                <th className="px-4 py-3 font-medium">{t("users.col.id")}</th>
                <th className="px-4 py-3 font-medium">{t("users.col.username")}</th>
                <th className="px-4 py-3 font-medium">{t("users.col.email")}</th>
                <th className="px-4 py-3 font-medium">{t("users.col.department")}</th>
                <th className="px-4 py-3 font-medium">{t("users.col.role")}</th>
                <th className="px-4 py-3 font-medium">{t("users.col.createdAt")}</th>
                <th className="px-4 py-3 text-right font-medium">{t("users.col.actions")}</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id} className="border-t border-slate-100 dark:border-slate-800">
                  <td className="px-4 py-3 text-slate-700 dark:text-slate-200">{u.id}</td>
                  <td className="px-4 py-3 font-medium text-slate-800 dark:text-slate-100">
                    {u.username}
                  </td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{u.email}</td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                    {departments.find((d) => d.id === u.departmentId)?.name ?? t("common.empty")}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded px-2 py-0.5 text-xs ${
                        u.role === "admin"
                          ? "bg-indigo-50 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300"
                          : "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300"
                      }`}
                    >
                      {roleLabel(t, u.role)}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                    {formatDateTime(u.createdAt, locale)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openEdit(u)}
                      className="rounded px-2 py-1 text-slate-500 transition hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
                    >
                      {t("common.edit")}
                    </button>
                    <button
                      onClick={() => remove(u)}
                      className="ml-1 rounded px-2 py-1 text-rose-500 transition hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-500/15"
                    >
                      {t("common.delete")}
                    </button>
                  </td>
                </tr>
              ))}
              {users.length === 0 && (
                <tr>
                  <td
                    colSpan={7}
                    className="px-4 py-10 text-center text-slate-400 dark:text-slate-500"
                  >
                    {t("users.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <Modal
          open={open}
          title={editing ? t("users.edit") : t("users.create")}
          onClose={() => setOpen(false)}
          footer={
            <>
              <button
                onClick={() => setOpen(false)}
                className="rounded-lg border border-slate-200 px-4 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
              >
                {t("common.cancel")}
              </button>
              <button
                onClick={save}
                disabled={saving}
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700 disabled:opacity-60"
              >
                {saving ? t("common.saving") : t("common.save")}
              </button>
            </>
          }
        >
          <Field label={t("users.field.username")}>
            <input
              className={inputCls}
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
              placeholder={t("users.field.usernamePlaceholder")}
            />
          </Field>
          <Field label={t("users.field.department")}>
            <select
              className={inputCls}
              value={form.departmentId}
              onChange={(e) => setForm({ ...form, departmentId: e.target.value })}
            >
              <option value="">{t("users.unassigned")}</option>
              {departments.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name}
                </option>
              ))}
            </select>
          </Field>
          <Field label={t("users.field.email")}>
            <input
              type="email"
              className={`${inputCls} ${
                emailInvalid
                  ? "border-rose-400 focus:border-rose-400 focus:ring-rose-100 dark:border-rose-500/60 dark:focus:border-rose-400 dark:focus:ring-rose-500/20"
                  : ""
              }`}
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              placeholder="you@example.com"
            />
            {emailInvalid && (
              <p className="mt-1 text-xs text-rose-500 dark:text-rose-400">
                {t("validate.emailInvalid")}
              </p>
            )}
          </Field>
          <Field label={editing ? t("users.field.passwordKeep") : t("users.field.password")}>
            <input
              type="password"
              className={inputCls}
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              placeholder="******"
            />
          </Field>
          <Field label={t("users.field.role")}>
            <select
              className={inputCls}
              value={form.role}
              onChange={(e) => setForm({ ...form, role: e.target.value as Role })}
            >
              <option value="member">{t("role.member")}</option>
              <option value="admin">{t("role.admin")}</option>
            </select>
          </Field>
        </Modal>
      </main>
    </AuthGuard>
  );
}
