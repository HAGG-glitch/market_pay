import apiClient from "./client";
import type { User, UserRole } from "@/types";

interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

interface MeResponse {
  user_id: string;
  email?: string;
  phone?: string;
  role: UserRole;
  is_demo?: boolean;
  display_name?: string;
}

export async function login(email: string, password: string) {
  const { data: tokens } = await apiClient.post<TokenPair>("/auth/login", {
    email,
    password,
  });

  const { data: profile } = await apiClient.get<MeResponse>("/auth/me", {
    headers: { Authorization: `Bearer ${tokens.access_token}` },
  });

  const user: User = {
    id: profile.user_id,
    name: profile.display_name || email.split("@")[0],
    email: profile.email || email,
    phone: profile.phone || "",
    role: profile.role,
  };

  return {
    user,
    token: tokens.access_token,
    refreshToken: tokens.refresh_token,
  };
}

export async function refreshToken(token: string) {
  const { data } = await apiClient.post<TokenPair>("/auth/refresh", {
    refresh_token: token,
  });
  return { token: data.access_token, refreshToken: data.refresh_token };
}

export async function getProfile() {
  const { data } = await apiClient.get<MeResponse>("/auth/me");
  return {
    id: data.user_id,
    role: data.role,
  };
}

export async function logout() {
  await apiClient.post("/auth/logout");
}

export async function vendorLogin(phone: string, pin: string) {
  const { data: tokens } = await apiClient.post<TokenPair>(
    "/auth/vendor-login",
    { phone, pin }
  );

  const { data: profile } = await apiClient.get<MeResponse>("/auth/me", {
    headers: { Authorization: `Bearer ${tokens.access_token}` },
  });

  const user: User = {
    id: profile.user_id,
    name: profile.display_name || phone,
    email: profile.email || "",
    phone: profile.phone || phone,
    role: profile.role,
  };

  return {
    user,
    token: tokens.access_token,
    refreshToken: tokens.refresh_token,
  };
}
