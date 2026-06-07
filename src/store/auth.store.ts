import { create } from "zustand";
import type { User } from "@/types";

interface AuthStore {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  setAuth: (user: User, token: string, refreshToken: string) => void;
  logout: () => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  token: null,
  refreshToken: null,
  isAuthenticated: false,

  setAuth: (user, token, refreshToken) => {
    localStorage.setItem("marketpay_token", token);
    localStorage.setItem("marketpay_refresh", refreshToken);
    localStorage.setItem("marketpay_user", JSON.stringify(user));
    set({ user, token, refreshToken, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem("marketpay_token");
    localStorage.removeItem("marketpay_refresh");
    localStorage.removeItem("marketpay_user");
    set({ user: null, token: null, refreshToken: null, isAuthenticated: false });
  },

  hydrate: () => {
    const token = localStorage.getItem("marketpay_token");
    const refreshToken = localStorage.getItem("marketpay_refresh");
    const userStr = localStorage.getItem("marketpay_user");
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr) as User;
        set({ user, token, refreshToken, isAuthenticated: true });
      } catch {
        set({ user: null, token: null, refreshToken: null, isAuthenticated: false });
      }
    }
  },
}));
