import { pageMetadata } from "@/lib/i18n/server";

// 回收箱页标题
export function generateMetadata() {
  return pageMetadata("page.trash");
}

export default function TrashLayout({ children }: { children: React.ReactNode }) {
  return children;
}
