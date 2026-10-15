import { API_BASE, fileUrl } from "@/lib/api";

/** 后端 API 的源（去掉 /api），用于把相对附件地址补全 */
const ORIGIN = API_BASE.replace(/\/api\/?$/, "");

/** FILE_PREFIX 附件对外地址前缀，与后端 storage.URLPrefix 保持一致 */
const FILE_PREFIX = "/api/files/";

/** mapUrls 只改 src / href 属性值，避免误伤正文里出现的同名字符串 */
function mapUrls(html: string, fn: (url: string) => string): string {
  return html.replace(
    /(\s(?:src|href)\s*=\s*")([^"]*)(")/gi,
    (_m, pre: string, url: string, post: string) => pre + fn(url) + post,
  );
}

/**
 * 富文本里存的是相对地址（/api/files/xxx），渲染 / 编辑前补全成后端地址
 * （带上下载票据），否则浏览器会按前端站点解析，图片 404。
 */
export function resolveFileUrls(html: string): string {
  if (!html) return "";
  return mapUrls(html, (url) => (url.startsWith(FILE_PREFIX) ? fileUrl(url) : url));
}

/** stripQuery 去掉附件地址上附带的下载票据等查询参数，只保留纯文件名地址 */
function stripQuery(rel: string): string {
  const i = rel.indexOf("?");
  return i === -1 ? rel : rel.slice(0, i);
}

/**
 * 保存前把绝对附件地址还原成相对地址（并剥掉下载票据），便于换域名 / 端口；
 * 后端富文本净化的 img src 只接受纯 /api/files/<随机名>，带查询串会被整条丢弃。
 */
export function toStoredHtml(html: string): string {
  if (!html) return "";
  const prefix = `${ORIGIN}${FILE_PREFIX}`;
  return mapUrls(html, (url) => {
    if (url.startsWith(prefix)) return stripQuery(url.slice(ORIGIN.length));
    if (url.startsWith(FILE_PREFIX)) return stripQuery(url);
    return url;
  });
}

/**
 * replaceFileTicket 把富文本里附件地址上的下载票据换成新的：
 * 票据首次取回或续期后，已渲染出来的图片地址跟着更新（否则仍是旧票据 / 无票据）。
 */
export function replaceFileTicket(html: string, ticket: string): string {
  if (!html || !ticket) return html;
  return mapUrls(html, (url) => {
    if (!url.startsWith(ORIGIN + FILE_PREFIX)) return url;
    return `${stripQuery(url)}?ticket=${encodeURIComponent(ticket)}`;
  });
}

/** 富文本转纯文本（卡片摘要等纯文本场景） */
export function richTextToText(html: string): string {
  if (!html) return "";
  if (typeof document === "undefined") {
    return html.replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();
  }
  const box = document.createElement("div");
  box.innerHTML = html;
  return (box.textContent ?? "").replace(/\s+/g, " ").trim();
}

/** 判断富文本是否为空（编辑器删干净后会残留 <p><br></p> 之类） */
export function isRichTextEmpty(html: string): boolean {
  if (!html) return true;
  if (/<img[\s>]/i.test(html)) return false;
  const text = html
    .replace(/<br\s*\/?>/gi, "")
    .replace(/&nbsp;/gi, " ")
    .replace(/<[^>]*>/g, "")
    .trim();
  return text === "";
}

/** 附件大小展示 */
export function formatBytes(size: number): string {
  if (!size || size < 0) return "";
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}
