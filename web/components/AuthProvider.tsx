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
  api,
  clearFileTicket,
  ensureFileTicket,
  getToken,
  setToken,
} from "@/lib/api";
import type { User } from "@/lib/types";

interface AuthState {
  user: User | null;
  loading: boolean;
  /** 附件下载票据：值变化时消费方重渲染，从而把新票据带进取回的下载地址 */
  fileTicket: string;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [fileTicket, setFileTicket] = useState("");
  const [loading, setLoading] = useState(true);

  // syncTicket 换取附件下载票据；失败不影响登录态，只是附件 / 图片暂时取不到
  const syncTicket = useCallback(async () => {
    setFileTicket(await ensureFileTicket());
  }, []);

  const refresh = useCallback(async () => {
    if (!getToken()) {
      clearFileTicket();
      setFileTicket("");
      setUser(null);
      setLoading(false);
      return;
    }
    try {
      setUser(await api.me());
    } catch {
      setToken(null);
      clearFileTicket();
      setFileTicket("");
      setUser(null);
    } finally {
      setLoading(false);
    }
    await syncTicket();
  }, [syncTicket]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  // 下载票据是短时效的：每分钟检查一次，临近过期时静默换新
  useEffect(() => {
    if (!user) return;
    const timer = setInterval(() => {
      void syncTicket();
    }, 60 * 1000);
    return () => clearInterval(timer);
  }, [user, syncTicket]);

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await api.login(email, password);
      setToken(res.token);
      clearFileTicket(); // 换账号必须丢弃上一张票据
      setUser(res.user);
      setLoading(false);
      await syncTicket();
    },
    [syncTicket],
  );

  const logout = useCallback(async () => {
    try {
      await api.logout();
    } catch {
      /* ignore */
    }
    setToken(null);
    clearFileTicket();
    setFileTicket("");
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({ user, loading, fileTicket, login, logout, refresh }),
    [user, loading, fileTicket, login, logout, refresh],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth 必须在 AuthProvider 内使用");
  return ctx;
}
