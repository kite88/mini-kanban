import { pageMetadata } from "@/lib/i18n/server";

// 看板标题：不跟项目名走（服务端渲染模式下这里拿不到登录态，也拿不到项目数据）
export function generateMetadata() {
  return pageMetadata("page.board");
}

export default function BoardLayout({ children }: { children: React.ReactNode }) {
  return children;
}
