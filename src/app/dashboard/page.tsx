"use client";

import { useAuthStore } from "@/store/auth.store";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { UserRole } from "@/types";

function roleToRoute(role: UserRole): string {
  return role.toLowerCase().replace(/_/g, "-");
}

export default function DashboardPage() {
  const { user } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (user) {
      router.replace(`/dashboard/${roleToRoute(user.role)}`);
    }
  }, [user, router]);

  return (
    <div className="flex h-screen items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-[#486B6D] border-t-transparent" />
    </div>
  );
}
