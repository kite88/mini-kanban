"use client";

import { useRef, useState } from "react";
import { useI18n } from "@/components/I18nProvider";
import { api, fileUrl } from "@/lib/api";
import { formatBytes } from "@/lib/html";
import type { Attachment } from "@/lib/types";

// 任务附件：支持多选上传、逐个移除；只读时仅展示下载链接
export function AttachmentList({
  items,
  onChange,
  readOnly,
}: {
  items: Attachment[];
  onChange: (v: Attachment[]) => void;
  readOnly?: boolean;
}) {
  const { t } = useI18n();
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");

  const add = async (files: File[]) => {
    const list = files;
    if (list.length === 0) return;
    setError("");
    setUploading(true);
    try {
      const uploaded = await api.uploadFiles(list);
      const merged = [...items];
      for (const a of uploaded) {
        if (!merged.some((x) => x.url === a.url)) merged.push(a);
      }
      onChange(merged);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("common.uploadFailed"));
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="space-y-1.5">
      {items.length === 0 && !readOnly && (
        <p className="text-xs text-slate-400 dark:text-slate-500">{t("attachment.empty")}</p>
      )}
      {items.length > 0 && (
        <ul className="space-y-1">
          {items.map((a) => (
            <li
              key={a.url}
              className="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-sm dark:border-slate-700 dark:bg-slate-900"
            >
              <a
                href={fileUrl(a.url)}
                target="_blank"
                rel="noreferrer"
                title={a.name}
                className="flex-1 truncate text-indigo-600 hover:underline dark:text-indigo-400"
              >
                {a.name}
              </a>
              <span className="shrink-0 text-xs text-slate-400 dark:text-slate-500">
                {formatBytes(a.size)}
              </span>
              {!readOnly && (
                <button
                  type="button"
                  title={t("attachment.removeTitle")}
                  onClick={() => onChange(items.filter((x) => x.url !== a.url))}
                  className="shrink-0 rounded px-1.5 text-xs text-rose-500 transition hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-500/15"
                >
                  {t("attachment.remove")}
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
      {!readOnly && (
        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            disabled={uploading}
            onClick={() => fileRef.current?.click()}
            className="rounded-lg border border-dashed border-slate-300 px-3 py-1.5 text-xs text-slate-500 transition hover:border-indigo-400 hover:text-indigo-600 disabled:opacity-60 dark:border-slate-600 dark:text-slate-400 dark:hover:border-indigo-500 dark:hover:text-indigo-400"
          >
            {uploading ? t("attachment.uploading") : t("attachment.add")}
          </button>
          <span className="text-xs text-slate-400 dark:text-slate-500">
            {t("attachment.hint")}
          </span>
        </div>
      )}
      <input
        ref={fileRef}
        type="file"
        multiple
        className="hidden"
        onChange={(e) => {
          // 必须先快照成数组：FileList 是实时的，清空 value 后它就空了
          const files = Array.from(e.target.files ?? []);
          e.target.value = "";
          void add(files);
        }}
      />
      {error && <p className="text-xs text-rose-600 dark:text-rose-400">{error}</p>}
    </div>
  );
}
