import type { Metadata } from "next";
import { cookies } from "next/headers";
import {
  DEFAULT_LANG,
  LANG_COOKIE,
  normalizeLang,
  type Lang,
  type MessageKey,
} from "./index";
import { pageTitle } from "./pageTitle";

/** 与 layout.tsx、I18nProvider 用同一个口径判断「是不是静态导出构建」 */
const STATIC_EXPORT = process.env.NEXT_PUBLIC_STATIC === "1";

/**
 * 服务端读语言 cookie —— 服务端渲染模式下首屏文案（含页面标题）的唯一口径。
 * 客户端 I18nProvider 拿到的是同一个值，所以不会出现英文界面闪一下中文。
 * 静态导出模式没有请求上下文，只能落回默认语言（挂载后由客户端纠正）。
 */
export function currentLang(): Lang {
  if (STATIC_EXPORT) return DEFAULT_LANG;
  return normalizeLang(cookies().get(LANG_COOKIE)?.value) ?? DEFAULT_LANG;
}

/** 路由段 layout 的 generateMetadata 返回值：把「页面名 - {app.title}」交给 Next 写进 <title> */
export function pageMetadata(key: MessageKey): Metadata {
  return { title: pageTitle(currentLang(), key) };
}
