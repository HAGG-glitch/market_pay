"use client";

import { useState } from "react";
import { useLoans } from "@/hooks/use-loans";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { formatCurrency, formatDate } from "@/lib/utils";
import Link from "next/link";
import { Plus } from "lucide-react";

const statusColors: Record<string, "success" | "warning" | "danger" | "info" | "default"> = {
  DRAFT: "default",
  PENDING_REVIEW: "warning",
  UNDER_REVIEW: "warning",
  APPROVED: "info",
  REJECTED: "danger",
  DISBURSED: "info",
  ACTIVE: "success",
  CLOSED: "success",
  DEFAULTED: "danger",
};

export default function LoansPage() {
  const [statusFilter, setStatusFilter] = useState("PENDING_REVIEW");
  const { data: loansData } = useLoans({ status: statusFilter });
  const loans = loansData?.data || [];

  const stateTabs = [
    { label: "Pending Review", value: "PENDING_REVIEW" },
    { label: "Approved", value: "APPROVED" },
    { label: "Disbursed", value: "DISBURSED" },
    { label: "All", value: "" },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Loans</h1>
          <p className="text-gray-500">Track all loan applications</p>
        </div>
        <Link href="/loans/apply">
          <Button aria-label="Apply for a new loan">
            <Plus size={16} className="mr-1.5" aria-hidden="true" />
            Apply for Loan
          </Button>
        </Link>
      </div>

      <div className="flex gap-2">
        {stateTabs.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setStatusFilter(tab.value)}
            className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
              statusFilter === tab.value
                ? "bg-primary text-white"
                : "bg-gray-100 text-gray-600 hover:bg-gray-200"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Loans</CardTitle>
        </CardHeader>
        <CardContent>
          {loans.length === 0 ? (
            <p className="text-sm text-gray-500">No loans found.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-gray-50 text-left text-gray-500">
                    <th className="pb-3 pt-3 pl-4 font-medium">ID</th>
                    <th className="pb-3 pt-3 font-medium">Amount</th>
                    <th className="pb-3 pt-3 font-medium">Interest</th>
                    <th className="pb-3 pt-3 font-medium">Status</th>
                    <th className="pb-3 pt-3 font-medium">Funded By</th>
                    <th className="pb-3 pt-3 pr-4 font-medium" />
                  </tr>
                </thead>
                <tbody>
                  {loans.map((loan, i) => (
                    <tr key={loan.id} className={`border-b last:border-0 transition-colors hover:bg-gray-100 ${i % 2 === 0 ? "bg-white" : "bg-gray-50/50"}`}>
                      <td className="py-3 pl-4 font-mono text-xs text-gray-900">
                        {loan.id.slice(0, 8)}...
                      </td>
                      <td className="py-3 font-medium">
                        {formatCurrency(loan.amount)}
                      </td>
                      <td className="py-3">{loan.interest_rate}%</td>
                      <td className="py-3">
                        <Badge variant={statusColors[loan.status] || "default"}>
                          {loan.status}
                        </Badge>
                      </td>
                      <td className="py-3 text-gray-500">
                        {loan.funded_by ? loan.funded_by.slice(0, 8) : "—"}
                      </td>
                      <td className="py-3 pr-4">
                        <Link
                          href={`/loans/${loan.id}`}
                          className="text-primary hover:underline font-medium"
                          aria-label={`View loan ${loan.id.slice(0, 8)}`}
                        >
                          View
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
