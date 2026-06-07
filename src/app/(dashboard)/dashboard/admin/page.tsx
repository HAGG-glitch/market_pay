"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useLoans } from "@/hooks/use-loans";
import { formatCurrency } from "@/lib/utils";

export default function AdminDashboard() {
  const { data: loansData } = useLoans();
  const loans = loansData?.data || [];
  const pendingReview = loans.filter((l) => l.status === "PENDING_REVIEW").length;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Admin Dashboard</h1>
        <p className="text-gray-500">Loan portfolio management</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Total Loans</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{loans.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Pending Review</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-yellow-600">{pendingReview}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Total Disbursed</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatCurrency(loans.reduce((s, l) => s + l.amount, 0))}</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
