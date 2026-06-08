"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { usePayments } from "@/hooks/use-payments";
import { formatCurrency, formatDate } from "@/lib/utils";
import Link from "next/link";

export default function CustomerDashboard() {
  const { data: paymentsData } = usePayments();
  const payments = paymentsData?.data || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">My Dashboard</h1>
        <p className="text-gray-500">Make payments and view history</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Total Payments
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{payments.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Total Spent
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {formatCurrency(
                payments
                  .filter((p) => p.status === "SUCCESS")
                  .reduce((s, p) => s + p.amount, 0)
              )}
            </p>
          </CardContent>
        </Card>
      </div>

      <Link href="/payments">
        <Button size="lg" className="w-full" aria-label="Go to payments page">
          Pay a Vendor
        </Button>
      </Link>

      <Card>
        <CardHeader>
          <CardTitle>Recent Payments</CardTitle>
        </CardHeader>
        <CardContent>
          {payments.length === 0 ? (
            <p className="text-sm text-gray-500">No payments yet.</p>
          ) : (
            <div className="space-y-2">
              {payments.slice(0, 5).map((p, i) => (
                <div
                  key={p.id}
                  className={`flex items-center justify-between rounded-lg border p-3 transition-colors hover:bg-gray-50 ${i % 2 === 0 ? "bg-white" : "bg-gray-50/50"}`}
                >
                  <div>
                    <p className="text-sm font-medium text-gray-900">
                      {formatCurrency(p.amount)}
                    </p>
                    <p className="text-xs text-gray-500">{formatDate(p.created_at)}</p>
                  </div>
                  <span
                    className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                      p.status === "SUCCESS"
                        ? "bg-green-100 text-green-700"
                        : p.status === "FAILED"
                        ? "bg-red-100 text-red-700"
                        : "bg-yellow-100 text-yellow-700"
                    }`}
                  >
                    {p.status}
                  </span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
