// 看板路由的服务端外壳：只负责声明静态导出需要的参数，真正的页面在 ProjectBoard.tsx（客户端组件）。
//
// 为什么需要这层：静态导出（output: "export"）要求动态段提供 generateStaticParams，
// 而该导出只能出现在服务端组件里 —— 页面本身是 "use client"，不能直接声明。
//
// 为什么占位参数是 "_"：Next 判断「有没有 generateStaticParams」的依据是
// 「预渲染出的路径条数是否大于 0」（见 next/dist/build/index.js 里的
// hasGenerateStaticParams = !!prerenderRoutes?.length），返回空数组会被判定为缺失并直接构建失败。
// 于是静态导出会额外预渲染一份占位页 /projects/_/ —— 页面内容对所有 id 都一样，
// 真实 id 的深链由静态服务器映射到这个占位页、由前端从地址栏解析，
// 见 server/webstatic.go 的 boardPlaceholderID / boardPlaceholderFile。
//
// 服务端渲染模式下不需要占位：dynamicParams = true，任意 id 都在请求时渲染。
import ProjectBoard from "./ProjectBoard";

// 与 layout.tsx、I18nProvider 用同一个口径判断「是不是静态导出构建」。
const STATIC_EXPORT = process.env.NEXT_PUBLIC_STATIC === "1";

export const dynamicParams = !STATIC_EXPORT;

export async function generateStaticParams() {
  // 占位 id 必须与 server/webstatic.go 的 boardPlaceholderID 一致
  return STATIC_EXPORT ? [{ id: "_" }] : [];
}

export default function Page() {
  return <ProjectBoard />;
}
