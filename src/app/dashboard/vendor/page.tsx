"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useLoans } from "@/hooks/use-loans";
import { formatCurrency } from "@/lib/utils";
import Link from "next/link";

export default function VendorDashboard() {
  const { data: loansData } = useLoans();
  const loans = loansData?.data || [];
  const activeLoan = loans.find((l) => l.status === "ACTIVE") || loans[0];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Vendor Dashboard</h1>
        <p className="text-gray-500">Manage your loans and payments</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Loan Balance</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {activeLoan ? formatCurrency(activeLoan.amount) : formatCurrency(0)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Loan Status</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-[#486B6D]">
              {activeLoan?.status || "N/A"}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Interest Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {activeLoan ? `${activeLoan.interest_rate}%` : "N/A"}
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="flex gap-3">
        <Link href="/loans/apply">
          <Button size="lg">Apply for Loan</Button>
        </Link>
        <Link href="/payments">
          <Button variant="outline" size="lg">View Payments</Button>
        </Link>
      </div>

      {activeLoan && (
        <Card>
          <CardHeader>
            <CardTitle>Repayment Schedule</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {activeLoan.repayment_schedule.map((s, i) => (
                <div
                  key={i}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div>
                    <p className="text-sm font-medium text-gray-900">
                      {formatCurrency(s.amount)}
                    </p>
                    <p className="text-xs text-gray-500">{s.due_date}</p>
                  </div>
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                      s.paid
                        ? "bg-green-100 text-green-800"
                        : "bg-yellow-100 text-yellow-800"
                    }`}
                  >
                    {s.paid ? "Paid" : "Pending"}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
