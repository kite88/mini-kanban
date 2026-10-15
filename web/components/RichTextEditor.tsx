"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useI18n } from "@/components/I18nProvider";
import { api, fileUrl } from "@/lib/api";
import { isRichTextEmpty, resolveFileUrls } from "@/lib/html";

// 简易富文本编辑器：基于 contentEditable + execCommand（无额外依赖）。
// 产出的是受控 HTML，服务端还会用白名单再净化一次（utils.SanitizeRichText）。
export function RichTextEditor({
  value,
  onChange,
  readOnly,
  placeholder,
  minHeight = 140,
}: {
  value: string;
  onChange: (html: string) => void;
  readOnly?: boolean;
  placeholder?: string;
  minHeight?: number;
}) {
  const { t } = useI18n();
  // 未显式传入时用当前语言的默认占位提示
  const placeholderText = placeholder ?? t("editor.placeholder");
  const ref = useRef<HTMLDivElement>(null);
  // 记录本编辑器最后一次产出的 HTML：外部值与之不同才回写 DOM，避免打字时光标跳回开头
  const emitted = useRef<string>(value);
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");
  const [linkOpen, setLinkOpen] = useState(false);
  const [linkUrl, setLinkUrl] = useState("");
  // 占位提示由状态控制：contentEditable 聚焦后浏览器会自己塞 <br>，
  // 用 CSS :empty 判断会立刻失效（有内容时又可能误判成空），所以不依赖伪类。
  const [empty, setEmpty] = useState(() => isRichTextEmpty(value));
  // 链接输入框打开前的选区（见 saveSelection / restoreSelection）
  const savedRange = useRef<Range | null>(null);

  // 初次挂载：把存储的相对地址补全后写入
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (!el.innerHTML) el.innerHTML = resolveFileUrls(value);
    setEmpty(isRichTextEmpty(el.innerHTML));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 外部改值（例如打开另一个任务）时同步；自己输入产生的变化不回写
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (value === emitted.current) return;
    if (document.activeElement === el) return;
    el.innerHTML = resolveFileUrls(value);
    emitted.current = value;
    setEmpty(isRichTextEmpty(value));
  }, [value]);

  const emit = useCallback(() => {
    const html = ref.current?.innerHTML ?? "";
    emitted.current = html;
    setEmpty(isRichTextEmpty(html));
    onChange(html);
  }, [onChange]);

  const exec = useCallback(
    (command: string, arg?: string) => {
      if (readOnly) return;
      ref.current?.focus();
      document.execCommand(command, false, arg);
      emit();
    },
    [emit, readOnly],
  );

  // 打开链接输入框时先把编辑器里的选区存下来：
  // 焦点转到输入框后选区会丢，execCommand 的插入点会落到编辑器开头（链接就跑到最前面了）
  const saveSelection = () => {
    const el = ref.current;
    const sel = window.getSelection();
    if (!el || !sel || sel.rangeCount === 0) return;
    const range = sel.getRangeAt(0);
    if (el.contains(range.commonAncestorContainer)) {
      savedRange.current = range.cloneRange();
    }
  };

  const restoreSelection = () => {
    const el = ref.current;
    if (!el) return;
    el.focus();
    const range = savedRange.current;
    if (!range) return;
    const sel = window.getSelection();
    if (!sel) return;
    sel.removeAllRanges();
    sel.addRange(range);
  };

  const insertImage = useCallback(
    async (file: File) => {
      setError("");
      setUploading(true);
      try {
        const [uploaded] = await api.uploadFiles([file]);
        if (uploaded) {
          // 用绝对地址插入，图片立即可见；保存时再还原成相对地址
          exec("insertImage", fileUrl(uploaded.url));
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : t("editor.imageUploadFailed"));
      } finally {
        setUploading(false);
      }
    },
    [exec, t],
  );

  // 支持直接粘贴截图
  const onPaste = (e: React.ClipboardEvent<HTMLDivElement>) => {
    if (readOnly) return;
    const image = Array.from(e.clipboardData?.files ?? []).find((f) => f.type.startsWith("image/"));
    if (image) {
      e.preventDefault();
      void insertImage(image);
    }
  };

  const applyLink = () => {
    const url = linkUrl.trim();
    setLinkOpen(false);
    setLinkUrl("");
    if (!url || readOnly) return;
    const full = /^(https?:)?\/\//i.test(url) ? url : `https://${url}`;

    // 关键：恢复打开输入框前的选区，否则链接会插到内容开头
    restoreSelection();
    const sel = window.getSelection();
    if (sel && sel.isCollapsed) {
      // 没有选中文本时，把地址本身变成链接插入到光标处
      document.execCommand("insertHTML", false, `<a href="${escapeAttr(full)}">${escapeHtml(full)}</a>&nbsp;`);
    } else {
      document.execCommand("createLink", false, full);
    }
    savedRange.current = null;
    emit();
  };

  const cancelLink = () => {
    setLinkOpen(false);
    setLinkUrl("");
    savedRange.current = null;
  };

  return (
    <div className="rounded-lg border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-900">
      {!readOnly && (
        <div className="flex flex-wrap items-center gap-0.5 border-b border-slate-200 px-2 py-1.5 dark:border-slate-700">
          <ToolButton label="B" title={t("editor.bold")} strong onClick={() => exec("bold")} />
          <ToolButton label="I" title={t("editor.italic")} italic onClick={() => exec("italic")} />
          <ToolButton
            label="U"
            title={t("editor.underline")}
            underline
            onClick={() => exec("underline")}
          />
          <ToolButton
            label="S"
            title={t("editor.strike")}
            strike
            onClick={() => exec("strikeThrough")}
          />
          <Sep />
          <ToolButton
            label="H"
            title={t("editor.heading")}
            onClick={() => exec("formatBlock", "h3")}
          />
          <ToolButton
            label="•"
            title={t("editor.unorderedList")}
            onClick={() => exec("insertUnorderedList")}
          />
          <ToolButton
            label="1."
            title={t("editor.orderedList")}
            onClick={() => exec("insertOrderedList")}
          />
          <ToolButton
            label="❝"
            title={t("editor.quote")}
            onClick={() => exec("formatBlock", "blockquote")}
          />
          <ToolButton
            label="</>"
            title={t("editor.code")}
            mono
            onClick={() => exec("formatBlock", "pre")}
          />
          <Sep />
          <ToolButton
            label="🔗"
            title={t("editor.link")}
            onClick={() => {
              saveSelection();
              setLinkUrl("");
              setLinkOpen(true);
            }}
          />
          <ToolButton
            label={uploading ? "…" : "🖼"}
            title={t("editor.image")}
            disabled={uploading}
            onClick={() => fileRef.current?.click()}
          />
          <ToolButton
            label="⌫"
            title={t("editor.removeFormat")}
            onClick={() => exec("removeFormat")}
          />
          <span className="ml-auto text-[11px] text-slate-400">
            {uploading ? t("editor.uploading") : ""}
          </span>
        </div>
      )}

      {linkOpen && !readOnly && (
        <div className="flex items-center gap-2 border-b border-slate-200 px-2 py-1.5 dark:border-slate-700">
          <input
            autoFocus
            value={linkUrl}
            onChange={(e) => setLinkUrl(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                applyLink();
              }
              if (e.key === "Escape") {
                e.preventDefault();
                cancelLink();
              }
            }}
            placeholder="https://example.com"
            className="flex-1 rounded border border-slate-200 bg-white px-2 py-1 text-xs text-slate-700 outline-none focus:border-indigo-400 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-200"
          />
          <button
            type="button"
            onClick={applyLink}
            className="rounded bg-indigo-600 px-2 py-1 text-xs font-medium text-white hover:bg-indigo-700"
          >
            {t("common.confirmAction")}
          </button>
          <button
            type="button"
            onClick={cancelLink}
            className="rounded border border-slate-200 px-2 py-1 text-xs text-slate-500 hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {t("common.cancel")}
          </button>
        </div>
      )}

      <div className="relative">
        {empty && (
          <p
            aria-hidden="true"
            className="pointer-events-none absolute left-3 right-3 top-2 text-sm text-slate-400 dark:text-slate-500"
          >
            {placeholderText}
          </p>
        )}
        <div
          ref={ref}
          contentEditable={!readOnly}
          suppressContentEditableWarning
          role="textbox"
          aria-multiline="true"
          onInput={emit}
          onBlur={emit}
          onPaste={onPaste}
          style={{ minHeight }}
          className={`rte px-3 py-2 text-sm text-slate-700 outline-none dark:text-slate-200 ${
            readOnly ? "cursor-default" : ""
          }`}
        />
      </div>

      <input
        ref={fileRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0];
          e.target.value = "";
          if (file) void insertImage(file);
        }}
      />

      {error && (
        <p className="border-t border-rose-100 px-3 py-1.5 text-xs text-rose-600 dark:border-rose-500/30 dark:text-rose-400">
          {error}
        </p>
      )}
    </div>
  );
}

function Sep() {
  return <span className="mx-1 h-4 w-px bg-slate-200 dark:bg-slate-700" />;
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function escapeAttr(url: string): string {
  return escapeHtml(url);
}

function ToolButton({
  label,
  title,
  onClick,
  strong,
  italic,
  underline,
  strike,
  mono,
  disabled,
}: {
  label: string;
  title: string;
  onClick: () => void;
  strong?: boolean;
  italic?: boolean;
  underline?: boolean;
  strike?: boolean;
  mono?: boolean;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      title={title}
      disabled={disabled}
      // 关键：mousedown 不抢焦点，否则选区丢失、execCommand 落空
      onMouseDown={(e) => e.preventDefault()}
      onClick={onClick}
      className={`grid h-7 min-w-7 place-items-center rounded px-1.5 text-xs text-slate-600 transition hover:bg-slate-100 disabled:opacity-50 dark:text-slate-300 dark:hover:bg-slate-800 ${
        strong ? "font-bold" : ""
      } ${italic ? "italic" : ""} ${underline ? "underline" : ""} ${strike ? "line-through" : ""} ${
        mono ? "font-mono" : ""
      }`}
    >
      {label}
    </button>
  );
}

/** 只读展示富文本内容（与编辑器同样的排版样式） */
export function RichTextView({ html, empty }: { html: string; empty?: string }) {
  const { t } = useI18n();
  const resolved = resolveFileUrls(html);
  if (!resolved) {
    return (
      <p className="text-sm text-slate-400 dark:text-slate-500">{empty ?? t("task.noDescription")}</p>
    );
  }
  return <div className="rte text-sm text-slate-700 dark:text-slate-200" dangerouslySetInnerHTML={{ __html: resolved }} />;
}
