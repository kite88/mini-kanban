import { translate, type Lang, type MessageKey } from "./index";

/**
 * 浏览器标签页标题：固定格式 `页面名 - {app.title}`（如「登录 - Mini-Kanban」）。
 *
 * 页面文件本身都是 "use client"，不能导出 metadata，所以页面名分两处声明、口径必须一致：
 * 1. 服务端：各路由段的 layout.tsx 用 pageMetadata() 声明（首屏就是对的，见 lib/i18n/server.ts）；
 * 2. 客户端：DocumentTitle 按本文件的 titleKeyForPath() 在跳转 / 切换语言时刷新
 *    （静态导出产物构建期只能按默认语言写死标题，靠这一步纠正）。
 */
export function titleKeyForPath(pathname: string): MessageKey | null {
  // 静态导出开了 trailingSlash，"/login/" 与 "/login" 要落到同一个键
  const path = pathname.replace(/\/+$/, "") || "/";
  if (path === "/login") return "page.login";
  if (path === "/projects") return "page.projects";
  // 看板是 /projects/[id]，标题不跟项目名走（服务端渲染模式下 layout 也拿不到登录态）
  if (path.startsWith("/projects/")) return "page.board";
  if (path === "/users") return "page.users";
  if (path === "/departments") return "page.departments";
  if (path === "/logs") return "page.logs";
  if (path === "/trash") return "page.trash";
  // 首页（按登录态跳转）等没有页面名的路由只有站点名
  return null;
}

/** 拼标题：没有页面名时只有 `{app.title}`（首页、404 这类路由） */
export function pageTitle(lang: Lang, key: MessageKey | null): string {
  if (!key) return translate(lang, "app.title");
  return `${translate(lang, key)} - ${translate(lang, "app.title")}`;
}
