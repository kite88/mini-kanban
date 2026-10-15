import type { Metadata } from "next";
import { cookies } from "next/headers";
import { AuthProvider } from "@/components/AuthProvider";
import { ConfirmProvider } from "@/components/ConfirmProvider";
import { I18nProvider } from "@/components/I18nProvider";
import { ThemeProvider, themeInitScript } from "@/components/ThemeProvider";
import {
  DEFAULT_LANG,
  LANG_COOKIE,
  langInitScript,
  normalizeLang,
  translate,
  type Lang,
} from "@/lib/i18n";
import "./globals.css";

/**
 * 静态导出模式（NEXT_OUTPUT=export）下没有请求上下文，取不到 cookie，
 * 首屏只能按默认语言渲染，挂载后由 I18nProvider 按 cookie / localStorage 纠正。
 * 这是纯静态产物换来的代价：非默认语言的首屏会闪一下（与 Vue dist 这类 SPA 一致）。
 */
const STATIC_EXPORT = process.env.NEXT_PUBLIC_STATIC === "1";

/**
 * 服务端读语言 cookie —— 服务端渲染模式下首屏文案的唯一口径。
 * 客户端 I18nProvider 拿到的是同一个值，所以 <html lang>、页面标题与正文文案
 * 在首屏就是对的，不会出现英文界面闪一下中文。
 * （代价是路由由静态预渲染变为按请求渲染，本项目前端本就依赖登录态，影响可忽略。）
 */
function currentLang(): Lang {
  if (STATIC_EXPORT) return DEFAULT_LANG;
  return normalizeLang(cookies().get(LANG_COOKIE)?.value) ?? DEFAULT_LANG;
}

export function generateMetadata(): Metadata {
  const lang = currentLang();
  return {
    title: translate(lang, "app.title"),
    description: translate(lang, "app.description"),
  };
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const lang = currentLang();
  return (
    // 下面的内联脚本会在水合前改写 lang 属性， suppressHydrationWarning 用来避免无意义的告警
    <html lang={lang} suppressHydrationWarning>
      <head>
        {/* 在水合前应用已保存的主题，避免首屏闪烁 */}
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
        {/* 静态导出模式下按 cookie 纠正 <html lang>；SSR 模式下执行也无害（值一致） */}
        <script dangerouslySetInnerHTML={{ __html: langInitScript }} />
      </head>
      <body className="min-h-screen">
        {/* I18nProvider 放在最外层：内部所有组件（含 NavBar / 确认框）都能取到文案 */}
        <I18nProvider initialLang={lang}>
          <ThemeProvider>
            <AuthProvider>
              <ConfirmProvider>{children}</ConfirmProvider>
            </AuthProvider>
          </ThemeProvider>
        </I18nProvider>
      </body>
    </html>
  );
}
