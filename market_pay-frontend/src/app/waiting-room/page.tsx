"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth.store";
import { Clock, Shield, LogOut } from "lucide-react";

export default function WaitingRoomPage() {
  const { user, logout } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!user) {
      router.replace("/login");
    }
  }, [user, router]);

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  if (!user) return null;

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-amber-50 to-white p-4">
      <div className="w-full max-w-md text-center">
        <div className="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-amber-100">
          <Clock className="h-10 w-10 text-amber-600" />
        </div>

        <h1 className="mb-2 text-2xl font-bold text-gray-900">Account Pending Approval</h1>
        <p className="mb-2 text-gray-600">
          Thank you for registering with MarketPay, <span className="font-semibold">{user.name}</span>.
        </p>
        <p className="mb-8 text-sm text-gray-500">
          Your account is currently under review. A loan officer will verify your details and activate your account.
          You will be able to access full services once approved.
        </p>

        <div className="mb-8 rounded-xl border border-amber-200 bg-amber-50 p-6">
          <h2 className="mb-3 text-sm font-semibold text-amber-800">What happens next?</h2>
          <ol className="space-y-2 text-left text-sm text-amber-700">
            <li className="flex items-start gap-2">
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-amber-200 text-xs font-bold">1</span>
              <span>A loan officer reviews your registration</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-amber-200 text-xs font-bold">2</span>
              <span>Your identity and business details are verified</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-amber-200 text-xs font-bold">3</span>
              <span>You receive a notification once approved</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-amber-200 text-xs font-bold">4</span>
              <span>Log in again to access your full dashboard</span>
            </li>
          </ol>
        </div>

        <button
          onClick={handleLogout}
          className="inline-flex items-center gap-2 rounded-lg border border-gray-300 bg-white px-6 py-2.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
        >
          <LogOut size={16} />
          Sign Out
        </button>

        <div className="mt-8 flex items-center justify-center gap-2 text-xs text-gray-400">
          <Shield size={12} />
          Secured with end-to-end encryption
        </div>
      </div>
    </div>
  );
}
