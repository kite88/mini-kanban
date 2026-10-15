"use client";

import { useCallback, useEffect, useState } from "react";
import { AuthGuard } from "@/components/AuthGuard";
import { useAuth } from "@/components/AuthProvider";
import { useConfirm } from "@/components/ConfirmProvider";
import { useI18n } from "@/components/I18nProvider";
import { Field, Modal, inputCls } from "@/components/Modal";
import { api } from "@/lib/api";
import { formatDateTime } from "@/lib/date";
import { tNow } from "@/lib/i18n";
import type { Department } from "@/lib/types";

export default function DepartmentsPage() {
  const { user } = useAuth();
  const { t, locale } = useI18n();
  const confirm = useConfirm();
  const isAdmin = user?.role === "admin";

  const [departments, setDepartments] = useState<Department[]>([]);
  const [error, setError] = useState("");
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Department | null>(null);
  const [form, setForm] = useState({ name: "", remark: "" });
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      setDepartments(await api.listDepartments());
    } catch (err) {
      setError(err instanceof Error ? err.message : tNow("common.loadFailed"));
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setEditing(null);
    setForm({ name: "", remark: "" });
    setOpen(true);
  };

  const openEdit = (d: Department) => {
    setEditing(d);
    setForm({ name: d.name, remark: d.remark ?? "" });
    setOpen(true);
  };

  const save = async () => {
    const name = form.name.trim();
    const remark = form.remark.trim();
    if (!name) {
      setError(t("departments.nameRequired"));
      return;
    }
    setError("");
    setSaving(true);
    try {
      if (editing) await api.updateDepartment(editing.id, { name, remark });
      else await api.createDepartment({ name, remark });
      setOpen(false);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("common.saveFailed"));
    } finally {
      setSaving(false);
    }
  };

  const remove = async (d: Department) => {
    const ok = await confirm({
      title: t("departments.deleteTitle"),
      description: t("departments.deleteDesc", { name: d.name }),
      note: t("departments.deleteNote"),
      confirmText: t("common.delete"),
      danger: true,
    });
    if (!ok) return;
    try {
      await api.deleteDepartment(d.id);
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
              {t("departments.title")}
            </h1>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
              {t("departments.subtitle")}
            </p>
          </div>
          {isAdmin && (
            <button
              onClick={openCreate}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700"
            >
              {`+ ${t("departments.create")}`}
            </button>
          )}
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
                <th className="px-4 py-3 font-medium">{t("departments.col.id")}</th>
                <th className="px-4 py-3 font-medium">{t("departments.col.name")}</th>
                <th className="px-4 py-3 font-medium">{t("departments.col.remark")}</th>
                <th className="px-4 py-3 font-medium">{t("departments.col.userCount")}</th>
                <th className="px-4 py-3 font-medium">{t("departments.col.createdAt")}</th>
                {isAdmin && (
                  <th className="px-4 py-3 text-right font-medium">{t("departments.col.actions")}</th>
                )}
              </tr>
            </thead>
            <tbody>
              {departments.map((d) => (
                <tr key={d.id} className="border-t border-slate-100 dark:border-slate-800">
                  <td className="px-4 py-3 text-slate-700 dark:text-slate-200">{d.id}</td>
                  <td className="px-4 py-3 font-medium text-slate-800 dark:text-slate-100">
                    {d.name}
                  </td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                    {d.remark || t("common.empty")}
                  </td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{d.userCount ?? 0}</td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                    {formatDateTime(d.createdAt, locale)}
                  </td>
                  {isAdmin && (
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => openEdit(d)}
                        className="rounded px-2 py-1 text-slate-500 transition hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
                      >
                        {t("common.edit")}
                      </button>
                      <button
                        onClick={() => remove(d)}
                        className="ml-1 rounded px-2 py-1 text-rose-500 transition hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-500/15"
                      >
                        {t("common.delete")}
                      </button>
                    </td>
                  )}
                </tr>
              ))}
              {departments.length === 0 && (
                <tr>
                  <td
                    colSpan={isAdmin ? 6 : 5}
                    className="px-4 py-10 text-center text-slate-400 dark:text-slate-500"
                  >
                    {t("departments.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <Modal
          open={open}
          title={editing ? t("departments.edit") : t("departments.create")}
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
          <Field label={t("departments.field.name")}>
            <input
              className={inputCls}
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder={t("departments.field.namePlaceholder")}
              autoFocus
            />
          </Field>
          <Field label={t("departments.field.remark")}>
            <textarea
              className={inputCls}
              rows={3}
              value={form.remark}
              onChange={(e) => setForm({ ...form, remark: e.target.value })}
              placeholder={t("departments.field.remarkPlaceholder")}
            />
          </Field>
        </Modal>
      </main>
    </AuthGuard>
  );
}
