"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useLogin } from "@/hooks/use-auth";
import { useAuthStore } from "@/store/auth.store";
import { useModeStore } from "@/store/mode.store";
import { ModeToggle } from "@/components/mode-toggle";
import { useToast } from "@/components/toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Wallet, Shield, ChevronDown, ChevronUp } from "lucide-react";
import { UserRole } from "@/types";

function roleToRoute(role: UserRole): string {
  return role.toLowerCase().replace(/_/g, "-");
}

const demoAccounts: { role: UserRole; name: string; email: string }[] = [
  { role: UserRole.SUPER_ADMIN, name: "Super Admin Demo", email: "superadmin@marketpay.sl" },
  { role: UserRole.ADMIN, name: "Admin Demo", email: "admin.demo@marketpay.sl" },
  { role: UserRole.MFI_PARTNER, name: "MFI Demo", email: "mfi.demo@marketpay.sl" },
  { role: UserRole.LOAN_OFFICER, name: "Loan Officer Demo", email: "officer@marketpay.sl" },
  { role: UserRole.FIELD_AGENT, name: "Field Agent Demo", email: "agent@marketpay.sl" },
  { role: UserRole.VENDOR, name: "Vendor Demo", email: "vendor.demo@marketpay.sl" },
  { role: UserRole.CUSTOMER, name: "Customer Demo", email: "customer.demo@marketpay.sl" },
];

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showDev, setShowDev] = useState(false);
  const login = useLogin();
  const setAuth = useAuthStore((s) => s.setAuth);
  const { mode, hydrate: hydrateMode } = useModeStore();
  const router = useRouter();

  useEffect(() => {
    hydrateMode();
  }, [hydrateMode]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    login.mutate({ email, password });
  };

  const [loginMode, setLoginMode] = useState<"email" | "vendor">("email");
  const [phone, setPhone] = useState("");
  const [pin, setPin] = useState("");
  const [vendorError, setVendorError] = useState("");
  const { toast } = useToast();

  const quickLogin = (userEmail: string) => {
    setEmail(userEmail);
    setPassword("password");
    login.mutate({ email: userEmail, password: "password" });
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-sm">
        <div className="mb-6 flex justify-center">
          <ModeToggle />
        </div>

        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary">
            <Wallet className="h-8 w-8 text-white" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900">MarketPay</h1>
          <p className="mt-1 text-sm text-gray-500">
            {mode === "demo" ? "Demo environment — test data only" : "Live environment — production data"}
          </p>
        </div>

        <div className="mb-4 flex rounded-lg border p-1">
          <button
            type="button"
            onClick={() => setLoginMode("email")}
            className={`flex-1 rounded-md py-2 text-sm font-medium ${
              loginMode === "email" ? "bg-primary text-white" : "text-gray-600"
            }`}
          >
            Email Login
          </button>
          <button
            type="button"
            onClick={() => setLoginMode("vendor")}
            className={`flex-1 rounded-md py-2 text-sm font-medium ${
              loginMode === "vendor" ? "bg-primary text-white" : "text-gray-600"
            }`}
          >
            Vendor (Phone + PIN)
          </button>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault();
            setVendorError("");
            if (loginMode === "vendor") {
              import("@/lib/api/auth.service").then(({ vendorLogin }) => {
                vendorLogin(phone, pin)
                  .then((result) => {
                    setAuth(result.user, result.token, result.refreshToken);
                    sessionStorage.removeItem("marketpay_logout");
                    toast("Welcome! You are now signed in.", "success");
                    if (result.vendorStatus === "PENDING") {
                      router.push("/waiting-room");
                    } else {
                      router.push(`/dashboard/${roleToRoute(result.user.role)}`);
                    }
                  })
                  .catch((err: Error) => {
                    setVendorError(err.message || "Vendor login failed. Check phone (+232...) and PIN.");
                  });
              });
            } else {
              handleSubmit(e);
            }
          }}
          className="space-y-4"
        >
          {loginMode === "email" ? (
            <>
              <Input
                id="email"
                label="Email"
                type="email"
                placeholder="superadmin@marketpay.sl"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
              <Input
                id="password"
                label="Password"
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </>
          ) : (
            <>
                <Input
                  id="phone"
                  label="Phone"
                  type="tel"
                  placeholder="+23233346989"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  required
                />
              <Input
                id="pin"
                label="PIN"
                type="password"
                placeholder="4-digit PIN"
                value={pin}
                onChange={(e) => setPin(e.target.value)}
                maxLength={4}
                required
              />
            </>
          )}

          {loginMode === "email" && login.isError && (
            <p className="text-sm text-red-500">
              {login.error?.message || "Invalid credentials. Please try again."}
            </p>
          )}
          {loginMode === "vendor" && vendorError && (
            <p className="text-sm text-red-500">{vendorError}</p>
          )}

          <Button
            type="submit"
            className="w-full"
            size="lg"
            disabled={login.isPending}
          >
            {login.isPending ? "Signing in..." : "Sign In"}
          </Button>
        </form>

        <div className="mt-6 flex items-center justify-center gap-2 text-xs text-gray-400">
          <Shield size={12} />
          Secured with end-to-end encryption
        </div>

        {mode === "demo" && (
          <div className="mt-8 border-t pt-6">
            <button
              type="button"
              onClick={() => setShowDev(!showDev)}
              className="flex w-full items-center justify-center gap-2 rounded-lg border border-dashed border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-800"
            >
              {showDev ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
              {showDev ? "Hide" : "Show"} Quick Login (Demo Data)
            </button>

            {showDev && (
              <div className="mt-4 space-y-2">
                <p className="text-xs text-gray-500">
                  One-click login for all roles. Password: <strong>password</strong>
                </p>
                {demoAccounts.map((u) => (
                  <button
                    type="button"
                    key={u.email}
                    onClick={() => quickLogin(u.email)}
                    className="w-full rounded-lg border border-gray-200 px-4 py-2.5 text-left text-sm transition-colors hover:border-primary hover:bg-primary/5"
                  >
                    <span className="font-medium text-gray-900">{u.name}</span>
                    <span className="ml-2 rounded bg-amber-100 px-1.5 py-0.5 text-xs text-amber-800">
                      Demo Data
                    </span>
                    <span className="ml-2 text-xs text-gray-400">{u.role}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
