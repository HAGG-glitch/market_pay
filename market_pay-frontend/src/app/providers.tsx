"use client";

import { QueryProvider } from "@/providers/query-provider";
import { AuthHydrator } from "@/components/auth-hydrator";
import { ToastProvider } from "@/components/toast";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryProvider>
      <ToastProvider>
        <AuthHydrator />
        {children}
      </ToastProvider>
    </QueryProvider>
  );
}
