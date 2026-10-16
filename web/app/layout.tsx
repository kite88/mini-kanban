import type { Metadata } from "next";
import { AuthProvider } from "@/components/AuthProvider";
import { ConfirmProvider } from "@/components/ConfirmProvider";
import { DocumentTitle } from "@/components/DocumentTitle";
import { I18nProvider } from "@/components/I18nProvider";
import { ThemeProvider, themeInitScript } from "@/components/ThemeProvider";
import { langInitScript, translate } from "@/lib/i18n";
import { currentLang } from "@/lib/i18n/server";
import "./globals.css";

/**
 * 根 metadata 只负责站点名：具体页面（/login、/projects…）的标题由各自路由段的
 * layout.tsx 声明成「页面名 - {app.title}」，口径见 lib/i18n/pageTitle.ts。
 */
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
          {/* 按当前语言刷新标签页标题：静态导出构建期写死的是默认语言，靠它纠正 */}
          <DocumentTitle />
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
