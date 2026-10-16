import { pageMetadata } from "@/lib/i18n/server";

// 部门管理页标题
export function generateMetadata() {
  return pageMetadata("page.departments");
}

export default function DepartmentsLayout({ children }: { children: React.ReactNode }) {
  return children;
}
