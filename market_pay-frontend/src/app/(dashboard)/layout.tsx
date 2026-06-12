"use client";

import { useAuthStore } from "@/store/auth.store";
import { useModeStore } from "@/store/mode.store";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Sidebar } from "@/components/layout/sidebar";
import { ModeToggle } from "@/components/mode-toggle";
import { NotificationBell } from "@/components/notification-bell";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { isAuthenticated, user, hydrate } = useAuthStore();
  const hydrateMode = useModeStore((s) => s.hydrate);

  useEffect(() => {
    hydrate();
    hydrateMode();
    const loggedOut = sessionStorage.getItem("marketpay_logout") === "true";
    const mode = localStorage.getItem("marketpay_mode") || "demo";

    if (!isAuthenticated && !loggedOut) {
      if (mode === "live") {
        router.replace("/login");
      }
    }
  }, [isAuthenticated, hydrate, hydrateMode, router]);

  if (!user) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen bg-gray-50">
      <Sidebar />
      <main className="min-h-screen w-full pt-16 md:ml-64 md:pt-0">
        <div className="border-b bg-white px-4 py-2 md:px-8">
          <div className="mx-auto flex max-w-7xl items-center justify-end gap-3">
            <NotificationBell />
            <ModeToggle />
          </div>
        </div>
        <div className="mx-auto max-w-7xl px-4 py-6 md:px-8 animate-fade-in">
          {children}
        </div>
      </main>
    </div>
  );
}
