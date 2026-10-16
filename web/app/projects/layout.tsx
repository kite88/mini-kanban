import { pageMetadata } from "@/lib/i18n/server";

// 项目列表标题（子路由 /projects/[id] 有自己的 layout 覆盖它）
export function generateMetadata() {
  return pageMetadata("page.projects");
}

export default function ProjectsLayout({ children }: { children: React.ReactNode }) {
  return children;
}
