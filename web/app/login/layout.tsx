import { pageMetadata } from "@/lib/i18n/server";

// 页面是 "use client"（不能导出 metadata），标题在服务端外壳里声明成「登录 - {app.title}」；
// 路由 → 文案键的映射与客户端 DocumentTitle 共用，见 lib/i18n/pageTitle.ts
export function generateMetadata() {
  return pageMetadata("page.login");
}

export default function LoginLayout({ children }: { children: React.ReactNode }) {
  return children;
}
