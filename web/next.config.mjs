/** @type {import('next').NextConfig} */

// 两种产物由 NEXT_OUTPUT 决定，构建脚本负责传：
//   standalone（默认）→ .next/standalone，需要 Node 运行，保留服务端渲染的首屏语言
//   export           → out/ 纯静态文件，可交给任意静态服务器或由后端二进制内嵌
const output = process.env.NEXT_OUTPUT === "export" ? "export" : "standalone";

const nextConfig = {
  reactStrictMode: true,
  output,
  // 静态导出时不需要在客户端再导航一次，直接用生成的 .html
  // （standalone 模式下该字段无副作用）
  trailingSlash: output === "export",
};

export default nextConfig;
