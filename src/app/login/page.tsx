"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useLogin } from "@/hooks/use-auth";
import { useAuthStore } from "@/store/auth.store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Wallet, Shield, ChevronDown, ChevronUp } from "lucide-react";
import type { User } from "@/types";
import { UserRole } from "@/types";

function roleToRoute(role: UserRole): string {
  return role.toLowerCase().replace(/_/g, "-");
}

const devUsers: { role: UserRole; name: string; phone: string }[] = [
  { role: UserRole.SUPER_ADMIN, name: "Bola Admin", phone: "super_admin" },
  { role: UserRole.ADMIN, name: "Chioma Admin", phone: "admin" },
  { role: UserRole.LOAN_OFFICER, name: "Emeka Officer", phone: "loan_officer" },
  { role: UserRole.FIELD_AGENT, name: "Amina Agent", phone: "field_agent" },
  { role: UserRole.VENDOR, name: "Segun Vendor", phone: "vendor" },
  { role: UserRole.CUSTOMER, name: "Grace Customer", phone: "customer" },
  { role: UserRole.MFI_PARTNER, name: "Mr. Partner", phone: "mfi_partner" },
];

export default function LoginPage() {
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [showDev, setShowDev] = useState(false);
  const login = useLogin();
  const setAuth = useAuthStore((s) => s.setAuth);
  const router = useRouter();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    login.mutate({ phone, password });
  };

  const devLogin = (role: UserRole, name: string) => {
    const user: User = {
      id: `${role.toLowerCase()}_001`,
      name,
      email: `${role.toLowerCase()}@marketpay.local`,
      phone,
      role,
    };
    setAuth(user, "dev_token", "dev_refresh");
    router.push(`/dashboard/${roleToRoute(role)}`);
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary">
            <Wallet className="h-8 w-8 text-white" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900">MarketPay</h1>
          <p className="mt-1 text-sm text-gray-500">
            Sign in to your dashboard
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            id="phone"
            label="Phone Number"
            type="tel"
            placeholder="+234 800 000 0000"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            required
          />

          <Input
            id="password"
            label="PIN / Password"
            type="password"
            placeholder="Enter your PIN or password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          {login.isError && (
            <p className="text-sm text-red-500">
              {login.error?.message || "Invalid credentials. Please try again."}
            </p>
          )}

          <Button
            type="submit"
            className="w-full"
            size="lg"
            disabled={login.isPending}
            aria-label="Sign in to your account"
          >
            {login.isPending ? "Signing in..." : "Sign In"}
          </Button>
        </form>

        <div className="mt-6 flex items-center justify-center gap-2 text-xs text-gray-400">
          <Shield size={12} />
          Secured with end-to-end encryption
        </div>

        <div className="mt-8 border-t pt-6">
          <button
            type="button"
            onClick={() => setShowDev(!showDev)}
            className="flex w-full items-center justify-center gap-2 rounded-lg border border-dashed border-gray-300 px-4 py-3 text-sm text-gray-500 transition-colors hover:border-primary hover:text-primary"
          >
            {showDev ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
            {showDev ? "Hide" : "Show"} Developer Test Access
          </button>

          {showDev && (
            <div className="mt-4 space-y-2">
              <p className="text-xs text-gray-500">
                Click a role to instantly log in (no backend needed):
              </p>
              {devUsers.map((u) => (
                <button
                  type="button"
                  key={u.role}
                  onClick={() => devLogin(u.role, u.name)}
                  className="w-full rounded-lg border border-gray-200 px-4 py-2.5 text-left text-sm transition-colors hover:border-primary hover:bg-primary/5"
                >
                  <span className="font-medium text-gray-900">{u.name}</span>
                  <span className="ml-2 rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-500">
                    {u.role}
                  </span>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
