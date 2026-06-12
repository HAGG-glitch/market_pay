"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth.store";

export default function Home() {
  const router = useRouter();
  const { isAuthenticated, hydrate } = useAuthStore();

  useEffect(() => {
    hydrate();
    const loggedOut = sessionStorage.getItem("marketpay_logout") === "true";
    if (loggedOut || !isAuthenticated) {
      router.replace("/login");
      return;
    }
    router.replace("/dashboard");
  }, [isAuthenticated, hydrate, router]);

  return (
    <div className="flex h-screen items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
    </div>
  );
}
