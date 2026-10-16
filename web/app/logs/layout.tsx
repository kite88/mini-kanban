import { pageMetadata } from "@/lib/i18n/server";

// 操作日志页标题
export function generateMetadata() {
  return pageMetadata("page.logs");
}

export default function LogsLayout({ children }: { children: React.ReactNode }) {
  return children;
}
