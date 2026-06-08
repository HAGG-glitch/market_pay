import apiClient from "./client";
import type { User } from "@/types";

export async function login(phone: string, password: string) {
  const { data } = await apiClient.post("/auth/login", { phone, password });
  return data as { user: User; token: string; refreshToken: string };
}

export async function refreshToken(token: string) {
  const { data } = await apiClient.post("/auth/refresh", {
    refresh_token: token,
  });
  return data as { token: string };
}

export async function getProfile() {
  const { data } = await apiClient.get("/auth/profile");
  return data as User;
}
