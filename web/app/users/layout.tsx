import { pageMetadata } from "@/lib/i18n/server";

// 用户管理页标题
export function generateMetadata() {
  return pageMetadata("page.users");
}

export default function UsersLayout({ children }: { children: React.ReactNode }) {
  return children;
}
