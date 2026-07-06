"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState, useMemo } from "react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/store/auth.store";
import { useNotifications } from "@/hooks/use-notifications";
import {
  LayoutDashboard,
  Wallet,
  CreditCard,
  Users,
  BarChart3,
  FileText,
  LogOut,
  Menu,
  X,
  ClipboardList,
} from "lucide-react";
import type { UserRole } from "@/types";
import { Button } from "@/components/ui/button";

interface NavItem {
  label: string;
  href: string;
  icon: React.ReactNode;
}

const roleNavItems: Record<UserRole, NavItem[]> = {
  SUPER_ADMIN: [
    { label: "Dashboard", href: "/dashboard/super-admin", icon: <LayoutDashboard size={18} /> },
    { label: "Loans", href: "/loans", icon: <CreditCard size={18} /> },
    { label: "Payment Plans", href: "/payment-plans", icon: <ClipboardList size={18} /> },
    { label: "Payments", href: "/payments", icon: <Wallet size={18} /> },
    { label: "Vendors", href: "/vendors", icon: <Users size={18} /> },
    { label: "Groups", href: "/group-lending", icon: <Users size={18} /> },
    { label: "Analytics", href: "/analytics", icon: <BarChart3 size={18} /> },
    { label: "Audit Logs", href: "/audit-logs", icon: <FileText size={18} /> },
  ],
  ADMIN: [
    { label: "Dashboard", href: "/dashboard/admin", icon: <LayoutDashboard size={18} /> },
    { label: "Loans", href: "/loans", icon: <CreditCard size={18} /> },
    { label: "Payment Plans", href: "/payment-plans", icon: <ClipboardList size={18} /> },
    { label: "Payments", href: "/payments", icon: <Wallet size={18} /> },
    { label: "Vendors", href: "/vendors", icon: <Users size={18} /> },
    { label: "Groups", href: "/group-lending", icon: <Users size={18} /> },
    { label: "Analytics", href: "/analytics", icon: <BarChart3 size={18} /> },
  ],
  LOAN_OFFICER: [
    { label: "Dashboard", href: "/dashboard/loan-officer", icon: <LayoutDashboard size={18} /> },
    { label: "Loan Queue", href: "/loans", icon: <FileText size={18} /> },
    { label: "Applications", href: "/loans/apply", icon: <CreditCard size={18} /> },
    { label: "Payment Plans", href: "/payment-plans", icon: <ClipboardList size={18} /> },
    { label: "Vendors", href: "/vendors", icon: <Users size={18} /> },
    { label: "Groups", href: "/group-lending", icon: <Users size={18} /> },
  ],
  FIELD_AGENT: [
    { label: "Dashboard", href: "/dashboard/field-agent", icon: <LayoutDashboard size={18} /> },
    { label: "Onboard Vendor", href: "/vendors/onboard", icon: <Users size={18} /> },
    { label: "Vendors", href: "/vendors", icon: <Users size={18} /> },
    { label: "Groups", href: "/group-lending", icon: <Users size={18} /> },
  ],
  VENDOR: [
    { label: "Dashboard", href: "/dashboard/vendor", icon: <LayoutDashboard size={18} /> },
    { label: "Apply Loan", href: "/loans/apply", icon: <CreditCard size={18} /> },
    { label: "Payments", href: "/payments", icon: <Wallet size={18} /> },
    { label: "Group", href: "/group-lending", icon: <Users size={18} /> },
  ],
  CUSTOMER: [
    { label: "Dashboard", href: "/dashboard/customer", icon: <LayoutDashboard size={18} /> },
    { label: "Pay Vendor", href: "/payments", icon: <Wallet size={18} /> },
    { label: "History", href: "/payments", icon: <FileText size={18} /> },
  ],
  MFI_PARTNER: [
    { label: "Dashboard", href: "/dashboard/mfi-partner", icon: <LayoutDashboard size={18} /> },
    { label: "Portfolio", href: "/analytics", icon: <BarChart3 size={18} /> },
    { label: "Vendors", href: "/vendors", icon: <Users size={18} /> },
  ],
};

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuthStore();
  const role = user?.role;
  const navItems = role ? roleNavItems[role] : [];
  const [mobileOpen, setMobileOpen] = useState(false);
  const { data: notifications = [] } = useNotifications();

  const unreadCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const n of notifications) {
      if (!n.is_read) {
        counts[n.event_type] = (counts[n.event_type] || 0) + 1;
      }
    }
    return counts;
  }, [notifications]);

  const badgeFor = (eventTypes: string[]) => {
    const total = eventTypes.reduce((sum, et) => sum + (unreadCounts[et] || 0), 0);
    return total > 0 ? total : undefined;
  };

  const handleLogout = async () => {
    try {
      const { logout: apiLogout } = await import("@/lib/api/auth.service");
      await apiLogout();
    } catch {
      /* proceed with local logout */
    }
    sessionStorage.setItem("marketpay_logout", "true");
    logout();
    router.push("/login");
  };

  const handleNav = (href: string) => {
    setMobileOpen(false);
  };

  return (
    <>
      {!mobileOpen && (
        <button
          onClick={() => setMobileOpen(true)}
          className="fixed left-4 top-4 z-50 flex h-10 w-10 items-center justify-center rounded-lg bg-white shadow-md md:hidden"
          aria-label="Open navigation menu"
        >
          <Menu size={20} />
        </button>
      )}

      {mobileOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/50 md:hidden"
          onClick={() => setMobileOpen(false)}
        />
      )}

      <aside
        className={cn(
          "fixed left-0 top-0 z-40 flex h-screen w-64 flex-col border-r border-gray-200 bg-white transition-transform duration-200",
          "-translate-x-full md:translate-x-0",
          mobileOpen && "translate-x-0"
        )}
      >
        <div className="flex h-16 items-center justify-between border-b px-6">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
              <Wallet className="h-5 w-5 text-white" />
            </div>
            <span className="text-xl font-bold text-primary">MarketPay</span>
          </div>
          <button
            onClick={() => setMobileOpen(false)}
            className="flex h-10 w-10 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 md:hidden"
            aria-label="Close navigation menu"
          >
            <X size={20} />
          </button>
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto p-4" aria-label="Main navigation">
          {navItems.map((item) => {
            const isGroups = item.label === "Groups" || item.label === "Group";

            const badgeEventTypes: Record<string, string[]> = {
              Loans: ["LoanRequested", "LoanDisbursed", "LoanDefaulted"],
              "Loan Queue": ["LoanRequested", "LoanDisbursed", "LoanDefaulted"],
              Vendors: ["VendorCreated", "VendorRegistered"],
            };
            const badge = badgeFor(badgeEventTypes[item.label] || []);

            return (
              <div key={item.href} className="relative">
                <Link
                  href={isGroups ? "#" : item.href}
                  onClick={(e) => {
                    if (isGroups) {
                      e.preventDefault();
                      alert("Group lending is coming soon!");
                    } else {
                      handleNav(item.href);
                    }
                  }}
                  aria-current={pathname === item.href ? "page" : undefined}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-200",
                    isGroups
                      ? "text-gray-400 cursor-not-allowed"
                      : pathname === item.href
                        ? "bg-primary/10 text-primary shadow-sm"
                        : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                  )}
                >
                  {item.icon}
                  <span className="flex-1">{item.label}</span>
                  {badge !== undefined && (
                    <span className="flex h-5 min-w-[20px] items-center justify-center rounded-full bg-red-500 px-1.5 text-[10px] font-bold text-white">
                      {badge > 9 ? "9+" : badge}
                    </span>
                  )}
                </Link>
                {isGroups && (
                  <span className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-medium text-amber-700">
                    Soon
                  </span>
                )}
              </div>
            );
          })}
        </nav>

        <div className="border-t p-4">
          <div className="mb-3 flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-sm font-medium text-white">
              {user?.name?.charAt(0) || "U"}
            </div>
            <div className="flex-1 truncate">
              <p className="text-sm font-medium text-gray-900">{user?.name}</p>
              <p className="text-xs text-gray-500">{user?.role}</p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start gap-2 text-red-600 hover:text-red-700"
            onClick={handleLogout}
          >
            <LogOut size={16} />
            Logout
          </Button>
        </div>
      </aside>
    </>
  );
}
