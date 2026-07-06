"use client";

import { Bell } from "lucide-react";
import { useNotifications, useMarkNotificationRead } from "@/hooks/use-notifications";
import type { InAppNotification } from "@/lib/api/notification.service";
import { useState } from "react";
import { useRouter } from "next/navigation";

const eventRoutes: Record<string, string> = {
  VendorCreated: "/vendors",
  VendorRegistered: "/vendors",
  LoanRequested: "/loans",
  LoanApplied: "/loans",
  LoanApproved: "/loans",
  LoanRejected: "/loans",
  LoanDisbursed: "/loans",
  RepaymentReceived: "/payments",
  PaymentReceived: "/payments",
  PaymentCompleted: "/payments",
  AccountFrozen: "/vendors",
  AccountUnfrozen: "/vendors",
  GroupCreated: "/group-lending",
  GroupFrozen: "/group-lending",
  LoanDefaulted: "/loans",
};

export function NotificationBell() {
  const router = useRouter();
  const { data: notifications = [] } = useNotifications();
  const markRead = useMarkNotificationRead();
  const [open, setOpen] = useState(false);
  const unread = notifications.filter((n) => !n.is_read).length;

  const handleClick = (n: InAppNotification) => {
    if (!n.is_read) markRead.mutate(n.id);
    setOpen(false);
    const href = eventRoutes[n.event_type] || null;
    if (href) router.push(href);
  };

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="relative rounded-lg p-2 text-gray-600 hover:bg-gray-100"
        aria-label="Notifications"
      >
        <Bell size={20} />
        {unread > 0 && (
          <span className="absolute right-1 top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white">
            {unread > 9 ? "9+" : unread}
          </span>
        )}
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute right-0 z-50 mt-2 w-80 rounded-lg border bg-white shadow-lg">
            <div className="border-b px-4 py-3 font-medium text-gray-900">
              Notifications
            </div>
            <div className="max-h-80 overflow-y-auto">
              {notifications.length === 0 ? (
                <p className="px-4 py-6 text-center text-sm text-gray-500">
                  No notifications yet
                </p>
              ) : (
                notifications.slice(0, 20).map((n) => (
                  <button
                    key={n.id}
                    type="button"
                    onClick={() => handleClick(n)}
                    className={`w-full border-b px-4 py-3 text-left text-sm hover:bg-gray-50 ${
                      n.is_read ? "bg-white" : "bg-primary/5"
                    }`}
                  >
                    <p className="font-medium text-gray-900">{n.title}</p>
                    <p className="text-gray-500">{n.body}</p>
                  </button>
                ))
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
