"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/AuthProvider";
import { useI18n } from "@/components/I18nProvider";

export default function Home() {
  const { user, loading } = useAuth();
  const { t } = useI18n();
  const router = useRouter();

  useEffect(() => {
    if (!loading) router.replace(user ? "/projects" : "/login");
  }, [loading, user, router]);

  return (
    <div className="grid min-h-screen place-items-center text-slate-500 dark:text-slate-400">
      {t("common.loading")}
    </div>
  );
}
