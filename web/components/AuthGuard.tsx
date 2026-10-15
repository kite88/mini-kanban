"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/AuthProvider";
import { useI18n } from "@/components/I18nProvider";
import { NavBar } from "@/components/NavBar";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();
  const { t } = useI18n();
  const router = useRouter();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!loading) {
      if (!user) router.replace("/login");
      else setReady(true);
    }
  }, [loading, user, router]);

  if (loading || !ready) {
    return (
      <div className="flex min-h-screen items-center justify-center text-slate-500 dark:text-slate-400">
        {t("common.loading")}
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <NavBar />
      {children}
    </div>
  );
}
