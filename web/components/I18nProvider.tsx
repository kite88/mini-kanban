"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  LANG_LOCALES,
  persistLang,
  readLangCookie,
  readStoredLang,
  setCurrentLang,
  translate,
  type Lang,
  type TFunction,
} from "@/lib/i18n";

interface I18nState {
  /** 当前语言（同时是 <html lang> 与发给后端的 Accept-Language） */
  lang: Lang;
  /** Intl / toLocaleString 用的 locale，如 zh-CN、en-US */
  locale: string;
  setLang: (lang: Lang) => void;
  t: TFunction;
}

const I18nContext = createContext<I18nState | null>(null);

/**
 * I18nProvider 全局语言：语种由根布局（服务端）从 cookie 里读出来传进来，
 * 因此首屏 SSR 与客户端首次渲染用的是同一个语言 —— 既不会一闪而过，
 * 也不存在水合前后文案不一致。
 *
 * 静态导出模式（NEXT_PUBLIC_STATIC=1）下服务端取不到 cookie，initialLang 是默认语言，
 * 挂载后按 cookie / localStorage 纠正一次；放在 effect 里做，不会造成水合不一致，
 * 代价是非默认语言的首屏会闪一下（纯静态产物的固有成本）。
 */
export function I18nProvider({
  initialLang,
  children,
}: {
  initialLang: Lang;
  children: React.ReactNode;
}) {
  const [lang, setLangState] = useState<Lang>(initialLang);

  // 静态导出模式：挂载后按 cookie / localStorage 纠正语言（只做一次）
  useEffect(() => {
    if (process.env.NEXT_PUBLIC_STATIC !== "1") return;
    const stored = readLangCookie() ?? readStoredLang();
    if (stored && stored !== initialLang) {
      setLangState(stored);
      setCurrentLang(stored);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // cookie 被手工清掉时，用 localStorage 里的镜像兜底，并写回 cookie
  useEffect(() => {
    if (readLangCookie()) return;
    const stored = readStoredLang();
    if (stored && stored !== lang) {
      setLangState(stored);
      persistLang(stored);
    }
  }, [lang]);

  // 同步给非组件代码（lib/api.ts 的请求头与兜底文案）与 <html lang>
  useEffect(() => {
    setCurrentLang(lang);
    document.documentElement.lang = lang;
  }, [lang]);

  const setLang = useCallback((next: Lang) => {
    setLangState(next);
    setCurrentLang(next);
    persistLang(next);
  }, []);

  const t = useCallback<TFunction>((key, params) => translate(lang, key, params), [lang]);

  const value = useMemo(
    () => ({ lang, locale: LANG_LOCALES[lang], setLang, t }),
    [lang, setLang, t],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18nState {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error("useI18n 必须在 I18nProvider 内使用");
  return ctx;
}
