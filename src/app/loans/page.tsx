"use client";

import { useLoans } from "@/hooks/use-loans";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatCurrency, formatDate } from "@/lib/utils";
import Link from "next/link";

const statusColors: Record<string, "success" | "warning" | "danger" | "info" | "default"> = {
  DRAFT: "default",
  PENDING_REVIEW: "warning",
  APPROVED: "info",
  DISBURSED: "info",
  ACTIVE: "success",
  CLOSED: "success",
  DEFAULTED: "danger",
};

export default function LoansPage() {
  const { data: loansData } = useLoans();
  const loans = loansData?.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Loans</h1>
          <p className="text-gray-500">Track all loan applications</p>
        </div>
        <Link
          href="/loans/apply"
          className="inline-flex h-10 items-center rounded-md bg-[#486B6D] px-4 text-sm font-medium text-white hover:bg-[#3a5a5c]"
        >
          Apply for Loan
        </Link>
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
                  <tr className="border-b text-left text-gray-500">
                    <th className="pb-3 font-medium">ID</th>
                    <th className="pb-3 font-medium">Amount</th>
                    <th className="pb-3 font-medium">Interest</th>
                    <th className="pb-3 font-medium">Status</th>
                    <th className="pb-3 font-medium">Funded By</th>
                    <th className="pb-3 font-medium" />
                  </tr>
                </thead>
                <tbody>
                  {loans.map((loan) => (
                    <tr key={loan.id} className="border-b last:border-0 hover:bg-gray-50">
                      <td className="py-3 font-mono text-xs text-gray-900">
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
                      <td className="py-3">
                        <Link
                          href={`/loans/${loan.id}`}
                          className="text-[#486B6D] hover:underline"
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
