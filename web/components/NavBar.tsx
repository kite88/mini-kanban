"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/components/AuthProvider";
import { useI18n } from "@/components/I18nProvider";
import { LanguageToggle } from "@/components/LanguageToggle";
import { ThemeToggle } from "@/components/ThemeToggle";
import { roleLabel } from "@/lib/labels";

export function NavBar() {
  const { user, logout } = useAuth();
  const { t } = useI18n();
  const pathname = usePathname();
  const router = useRouter();

  const links = [
    { href: "/projects", label: t("nav.projects") },
    // 回收箱对所有登录用户开放：普通用户只会看到自己有权处置的已删数据
    { href: "/trash", label: t("nav.trash") },
    ...(user?.role === "admin"
      ? [
          { href: "/users", label: t("nav.users") },
          { href: "/departments", label: t("nav.departments") },
          // 操作日志含用户 / 部门的敏感动作，与后端一致只对管理员开放
          { href: "/logs", label: t("nav.logs") },
        ]
      : []),
  ];

  return (
    <header className="sticky top-0 z-30 border-b border-slate-200 bg-white/80 backdrop-blur dark:border-slate-800 dark:bg-slate-900/80">
      <div className="mx-auto flex h-14 max-w-[1400px] items-center justify-between px-6">
        <div className="flex items-center gap-8">
          <span className="flex items-center gap-2 text-base font-semibold text-slate-900 dark:text-slate-100">
            <span className="grid h-7 w-7 place-items-center rounded-lg bg-indigo-600 text-xs text-white">
              K
            </span>
            {t("app.title")}
          </span>
          <nav className="flex items-center gap-1">
            {links.map((l) => (
              <Link
                key={l.href}
                href={l.href}
                className={`rounded-lg px-3 py-1.5 text-sm transition ${
                  pathname.startsWith(l.href)
                    ? "bg-indigo-50 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300"
                    : "text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
                }`}
              >
                {l.label}
              </Link>
            ))}
          </nav>
        </div>
        <div className="flex items-center gap-3 text-sm">
          <LanguageToggle />
          <ThemeToggle />
          <span className="text-slate-600 dark:text-slate-300">
            {user?.username}
            <span className="ml-2 rounded bg-slate-100 px-1.5 py-0.5 text-xs text-slate-500 dark:bg-slate-800 dark:text-slate-400">
              {user ? roleLabel(t, user.role) : ""}
            </span>
          </span>
          <button
            onClick={async () => {
              await logout();
              router.replace("/login");
            }}
            className="rounded-lg border border-slate-200 px-3 py-1.5 text-slate-600 transition hover:bg-slate-100 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {t("nav.logout")}
          </button>
        </div>
      </div>
    </header>
  );
}
