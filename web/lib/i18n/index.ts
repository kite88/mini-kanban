import { en } from "./en";
import { zhCN, type MessageKey } from "./zh-CN";

export type { MessageKey } from "./zh-CN";

/** 支持的语言（新增语言时：加一份字典 → 加进 DICTS / LANG_LABELS / LANG_LOCALES 即可） */
export type Lang = "zh-CN" | "en";

export const LANGS: Lang[] = ["zh-CN", "en"];

export const DEFAULT_LANG: Lang = "zh-CN";

/** 语言 cookie：服务端渲染时读它定 <html lang> 与页面标题，避免英文界面闪中文 */
export const LANG_COOKIE = "mini-kanban-lang";

/** localStorage 是 cookie 的镜像：cookie 被清掉时还能沿用上次的选择 */
export const LANG_STORAGE_KEY = "mini-kanban-lang";

/** 语言在界面上始终以它自己的语言呈现，不参与翻译（用于 title / aria-label） */
export const LANG_LABELS: Record<Lang, string> = {
  "zh-CN": "简体中文",
  en: "English",
};

/** 切换控件上的短标签：导航栏空间有限，用更短的两字 / 缩写 */
export const LANG_SHORT_LABELS: Record<Lang, string> = {
  "zh-CN": "中文",
  en: "EN",
};

/** Intl / toLocaleString 用的 locale（BCP-47），同时用于 <html lang> */
export const LANG_LOCALES: Record<Lang, string> = {
  "zh-CN": "zh-CN",
  en: "en-US",
};

export const DICTS: Record<Lang, Record<MessageKey, string>> = {
  "zh-CN": zhCN,
  en,
};

export type TParams = Record<string, string | number>;
export type TFunction = (key: MessageKey, params?: TParams) => string;

/** 把任意来源（cookie / localStorage / Accept-Language / html lang）的值收敛成受支持的语言 */
export function normalizeLang(value?: string | null): Lang | null {
  if (!value) return null;
  const v = value.trim().toLowerCase();
  if (!v) return null;
  // 精确命中
  const exact = LANGS.find((l) => l.toLowerCase() === v);
  if (exact) return exact;
  // 只取主子标签：zh、zh-hans、zh-cn-xxx → zh-CN；en-gb → en
  const primary = v.split(/[-_]/)[0];
  if (primary === "zh") return "zh-CN";
  if (primary === "en") return "en";
  return null;
}

/**
 * 取文案：找不到键时回退到源语言（中文），再找不到就原样返回键名，
 * 保证界面上永远有字，不会出现空白。
 */
export function translate(lang: Lang, key: MessageKey, params?: TParams): string {
  const dict = DICTS[lang] ?? DICTS[DEFAULT_LANG];
  const raw = dict[key] ?? DICTS[DEFAULT_LANG][key] ?? key;
  if (!params) return raw;
  return raw.replace(/\{(\w+)\}/g, (m, name: string) =>
    name in params ? String(params[name]) : m,
  );
}

// ---- 非组件代码（lib/ 的请求层、纯函数）用的当前语言 ----
// useI18n().t 供组件使用；组件渲染发生在服务端时模块级变量还是默认值，
// 所以绝不能拿 tNow 在 render 里取文案（会造成水合不一致）。
let currentLang: Lang = DEFAULT_LANG;

export function getCurrentLang(): Lang {
  return currentLang;
}

export function setCurrentLang(lang: Lang) {
  currentLang = lang;
}

/** 事件回调 / 异步流程里取文案：语言已在客户端生效，可放心使用 */
export function tNow(key: MessageKey, params?: TParams): string {
  return translate(currentLang, key, params);
}

// ---- 语言持久化（cookie 为主，localStorage 为镜像）----

function readCookie(name: string): string | null {
  if (typeof document === "undefined") return null;
  const prefix = `${name}=`;
  for (const part of document.cookie.split(";")) {
    const item = part.trim();
    if (item.startsWith(prefix)) return decodeURIComponent(item.slice(prefix.length));
  }
  return null;
}

export function readLangCookie(): Lang | null {
  return normalizeLang(readCookie(LANG_COOKIE));
}

/** 与主题一样写一年；HTTPS 下顺带加 Secure */
export function writeLangCookie(lang: Lang) {
  if (typeof document === "undefined") return;
  const secure = window.location.protocol === "https:" ? "; Secure" : "";
  document.cookie = `${LANG_COOKIE}=${lang}; path=/; max-age=31536000; SameSite=Lax${secure}`;
}

export function readStoredLang(): Lang | null {
  if (typeof window === "undefined") return null;
  try {
    return normalizeLang(window.localStorage.getItem(LANG_STORAGE_KEY));
  } catch {
    return null;
  }
}

export function writeStoredLang(lang: Lang) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(LANG_STORAGE_KEY, lang);
  } catch {
    /* 隐私模式下写入失败时忽略 */
  }
}

/** 切换语言：cookie + localStorage 一起写，服务端下次渲染就能读到 */
export function persistLang(lang: Lang) {
  writeLangCookie(lang);
  writeStoredLang(lang);
}

/**
 * 水合前同步 <html lang> 的内联脚本（与主题 themeInitScript 同一套路）。
 *
 * 只在静态导出模式下真正起作用：那时服务端不知道语言，HTML 里是默认语言，
 * 脚本先把 lang 属性改对，随后 I18nProvider 在挂载后把文案切过去。
 * 服务端渲染模式下语言由服务端决定，执行它结果一致、无副作用。
 */
export const langInitScript = `(function(){try{
  var m=document.cookie.match(/(?:^|; )${LANG_COOKIE}=([^;]*)/);
  var v=m?decodeURIComponent(m[1]):"";
  var p=(v||"").trim().toLowerCase().split(/[-_]/)[0];
  var l="${DEFAULT_LANG}";
  if(p==="en"){l="en";}else if(p==="zh"){l="zh-CN";}
  document.documentElement.lang=l;
}catch(e){}})();`;
