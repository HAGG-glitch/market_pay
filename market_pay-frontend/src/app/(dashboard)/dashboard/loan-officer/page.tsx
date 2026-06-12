"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useOfficerQueue, useDashboardSummary } from "@/hooks/use-reporting";
import { formatCurrency } from "@/lib/utils";

export default function LoanOfficerDashboard() {
  const { data: queue } = useOfficerQueue();
  const { data: summary } = useDashboardSummary();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Loan Officer Dashboard</h1>
        <p className="text-gray-500">Review queue and portfolio health</p>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Pending Review</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-amber-600">
              {queue?.pending_review ?? 0}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Under Review</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{queue?.under_review ?? 0}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Active Loans</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{summary?.active_loans ?? 0}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Overdue</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-red-600">
              {summary?.overdue_loans ?? 0}
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Portfolio Snapshot</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-gray-500">Outstanding Portfolio</p>
            <p className="text-xl font-bold">
              {formatCurrency(summary?.portfolio_value ?? 0)}
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Repayment Rate</p>
            <p className="text-xl font-bold text-green-600">
              {(summary?.repayment_rate ?? 0).toFixed(1)}%
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
