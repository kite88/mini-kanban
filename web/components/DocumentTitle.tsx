"use client";

import { useEffect } from "react";
import { usePathname } from "next/navigation";
import { useI18n } from "@/components/I18nProvider";
import { pageTitle, titleKeyForPath } from "@/lib/i18n/pageTitle";

/**
 * 标签页标题的第二道保险（服务端那道见各路由段的 layout.tsx）：
 * - 静态导出产物在构建期只能按默认语言写死标题，这里挂载后按真实语言纠正；
 * - 客户端路由跳转 / 切换语言时同步刷新（服务端渲染模式下两者值相同，无副作用）。
 */
export function DocumentTitle() {
  const pathname = usePathname();
  const { lang } = useI18n();

  useEffect(() => {
    document.title = pageTitle(lang, titleKeyForPath(pathname));
  }, [lang, pathname]);

  return null;
}
