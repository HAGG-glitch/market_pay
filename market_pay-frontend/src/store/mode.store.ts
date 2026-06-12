import { create } from "zustand";

export type AppMode = "demo" | "live";

interface ModeStore {
  mode: AppMode;
  setMode: (mode: AppMode) => void;
  hydrate: () => void;
}

export const useModeStore = create<ModeStore>((set) => ({
  mode: "demo",
  setMode: (mode) => {
    localStorage.setItem("marketpay_mode", mode);
    set({ mode });
  },
  hydrate: () => {
    const stored = localStorage.getItem("marketpay_mode") as AppMode | null;
    if (stored === "demo" || stored === "live") {
      set({ mode: stored });
    }
  },
}));

export function isDemoMode() {
  if (typeof window === "undefined") return true;
  return localStorage.getItem("marketpay_mode") !== "live";
}
