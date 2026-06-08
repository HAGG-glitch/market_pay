"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth.store";
import { UserRole } from "@/types";

export default function Home() {
  const router = useRouter();
  const { isAuthenticated, setAuth, hydrate } = useAuthStore();

  useEffect(() => {
    hydrate();
    const loggedOut = sessionStorage.getItem("marketpay_logout") === "true";

    if (loggedOut) {
      router.replace("/login");
      return;
    }

    if (!isAuthenticated) {
      setAuth(
        {
          id: "dev_001",
          name: "Dev Admin",
          email: "dev@marketpay.local",
          phone: "+2348000000000",
          role: UserRole.SUPER_ADMIN,
        },
        "dev_token",
        "dev_refresh"
      );
    }
    router.replace("/dashboard/super-admin");
  }, [isAuthenticated, setAuth, hydrate, router]);

  return (
    <div className="flex h-screen items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
    </div>
  );
}
