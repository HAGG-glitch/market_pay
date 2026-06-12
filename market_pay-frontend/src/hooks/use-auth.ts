"use client";

import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth.store";
import { login as loginApi } from "@/lib/api/auth.service";
import { UserRole } from "@/types";

function roleToRoute(role: UserRole): string {
  return role.toLowerCase().replace(/_/g, "-");
}

export function useLogin() {
  const setAuth = useAuthStore((s) => s.setAuth);
  const router = useRouter();

  return useMutation({
    mutationFn: ({
      email,
      password,
    }: {
      email: string;
      password: string;
    }) => loginApi(email, password),
    onSuccess: (data) => {
      setAuth(data.user, data.token, data.refreshToken);
      router.push(`/dashboard/${roleToRoute(data.user.role)}`);
    },
  });
}
