"use client";

import { useAuthStore } from "@/store/auth.store";
import { useEffect } from "react";
import { Sidebar } from "@/components/layout/sidebar";
import type { User, UserRole } from "@/types";

const devUser: User = {
  id: "dev_001",
  name: "Dev Admin",
  email: "dev@marketpay.local",
  phone: "+2348000000000",
  role: "SUPER_ADMIN" as UserRole,
};

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, user, setAuth, hydrate } = useAuthStore();

  useEffect(() => {
    hydrate();
    const loggedOut = sessionStorage.getItem("marketpay_logout") === "true";
    if (!isAuthenticated && !loggedOut) {
      setAuth(devUser, "dev_token", "dev_refresh");
    }
  }, [isAuthenticated, setAuth, hydrate]);

  if (!user) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-[#486B6D] border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen bg-gray-50">
      <Sidebar />
      <main className="min-h-screen w-full pt-16 md:ml-64 md:pt-0">
        <div className="mx-auto max-w-7xl px-4 py-6 md:px-8">
          {children}
        </div>
      </main>
    </div>
  );
}
