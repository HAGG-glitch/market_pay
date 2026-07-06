"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth.store";
import { LogOut, RefreshCw } from "lucide-react";

function decodeToken(token: string): { exp?: number } | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) return null;
    return JSON.parse(atob(parts[1]));
  } catch {
    return null;
  }
}

const INACTIVITY_TIMEOUT = 15 * 60 * 1000; // 15 min inactivity → logout
const WARN_BEFORE = 5 * 60 * 1000; // warn when 5 min remain

export function SessionTimer() {
  const router = useRouter();
  const logout = useAuthStore((s) => s.logout);
  const [timeLeft, setTimeLeft] = useState<number | null>(null);
  const [showWarning, setShowWarning] = useState(false);
  const inactivityRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const warnShownRef = useRef(false);

  const doLogout = useCallback(() => {
    logout();
    sessionStorage.setItem("marketpay_logout", "true");
    localStorage.removeItem("marketpay_token");
    localStorage.removeItem("marketpay_refresh");
    localStorage.removeItem("marketpay_user");
    router.push("/login");
  }, [logout, router]);

  const resetInactivity = useCallback(() => {
    if (inactivityRef.current) clearTimeout(inactivityRef.current);
    inactivityRef.current = setTimeout(doLogout, INACTIVITY_TIMEOUT);
  }, [doLogout]);

  const refreshSession = useCallback(() => {
    const refreshToken = localStorage.getItem("marketpay_refresh");
    if (!refreshToken) return;
    fetch(
      `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/auth/refresh`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      }
    )
      .then((r) => r.json())
      .then((data) => {
        if (data.access_token) {
          localStorage.setItem("marketpay_token", data.access_token);
          if (data.refresh_token) {
            localStorage.setItem("marketpay_refresh", data.refresh_token);
          }
          warnShownRef.current = false;
          setShowWarning(false);
          resetInactivity();
        }
      })
      .catch(() => {});
  }, [resetInactivity]);

  // Listen for user activity
  useEffect(() => {
    const events = ["mousedown", "keydown", "touchstart", "scroll"];
    const handler = () => resetInactivity();
    events.forEach((e) => window.addEventListener(e, handler));
    resetInactivity();
    return () => {
      events.forEach((e) => window.removeEventListener(e, handler));
      if (inactivityRef.current) clearTimeout(inactivityRef.current);
    };
  }, [resetInactivity]);

  // Tick the token timer
  useEffect(() => {
    const token = localStorage.getItem("marketpay_token");
    if (!token) return;

    const payload = decodeToken(token);
    if (!payload?.exp) return;

    const tick = () => {
      const now = Math.floor(Date.now() / 1000);
      const remaining = payload.exp! - now;
      setTimeLeft(remaining > 0 ? remaining : 0);

      if (remaining <= 0) {
        doLogout();
        return;
      }

      if (remaining <= 300 && !warnShownRef.current) {
        warnShownRef.current = true;
        setShowWarning(true);
      }

      // Try to refresh when < 2 min remain
      if (remaining <= 120) {
        const refreshToken = localStorage.getItem("marketpay_refresh");
        if (refreshToken) {
          fetch(
            `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/auth/refresh`,
            {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ refresh_token: refreshToken }),
            }
          )
            .then((r) => r.json())
            .then((data) => {
              if (data.access_token) {
                localStorage.setItem("marketpay_token", data.access_token);
                if (data.refresh_token) {
                  localStorage.setItem("marketpay_refresh", data.refresh_token);
                }
                warnShownRef.current = false;
                setShowWarning(false);
              }
            })
            .catch(() => {});
        }
      }
    };

    tick();
    const interval = setInterval(tick, 1000);
    return () => clearInterval(interval);
  }, [doLogout]);

  if (timeLeft === null) return null;

  const minutes = Math.floor(timeLeft / 60);
  const seconds = timeLeft % 60;
  const isUrgent = timeLeft <= 300;

  return (
    <>
      {showWarning && (
        <div className="fixed bottom-4 right-4 z-50 flex items-center gap-3 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 shadow-lg">
          <div className="flex-1">
            <p className="text-sm font-medium text-amber-800">
              Session expires in {minutes}:{seconds.toString().padStart(2, "0")}
            </p>
            <p className="text-xs text-amber-600">Continue working to extend</p>
          </div>
          <button
            onClick={refreshSession}
            className="flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-50"
          >
            <RefreshCw size={12} />
            Continue
          </button>
          <button
            onClick={doLogout}
            className="flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50"
          >
            <LogOut size={12} />
            Logout
          </button>
        </div>
      )}
      <div
        className={`flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ${
          isUrgent ? "bg-red-50 text-red-600" : "bg-gray-100 text-gray-500"
        }`}
        title="Session time remaining"
      >
        <span className="tabular-nums">
          {minutes}:{seconds.toString().padStart(2, "0")}
        </span>
      </div>
    </>
  );
}
