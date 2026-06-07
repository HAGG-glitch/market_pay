"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useLoans } from "@/hooks/use-loans";
import { usePayments } from "@/hooks/use-payments";
import { formatCurrency } from "@/lib/utils";

export default function SuperAdminDashboard() {
  const { data: loansData } = useLoans();
  const { data: paymentsData } = usePayments();
  const loans = loansData?.data || [];
  const payments = paymentsData?.data || [];

  const totalDisbursed = loans.reduce((s, l) => s + l.amount, 0);
  const totalRepaid = payments.filter((p) => p.status === "SUCCESS").reduce((s, p) => s + p.amount, 0);
  const activeLoans = loans.filter((l) => l.status === "ACTIVE").length;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Super Admin Dashboard</h1>
        <p className="text-gray-500">Full portfolio overview</p>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
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
            <CardTitle className="text-sm font-medium text-gray-500">Total Disbursed</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatCurrency(totalDisbursed)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Total Repaid</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatCurrency(totalRepaid)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Active Loans</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{activeLoans}</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Loans</CardTitle>
        </CardHeader>
        <CardContent>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-gray-500">
                <th className="pb-3 font-medium">ID</th>
                <th className="pb-3 font-medium">Amount</th>
                <th className="pb-3 font-medium">Status</th>
                <th className="pb-3 font-medium">Interest</th>
              </tr>
            </thead>
            <tbody>
              {loans.slice(0, 5).map((loan) => (
                <tr key={loan.id} className="border-b last:border-0">
                  <td className="py-3 text-gray-900">{loan.id.slice(0, 8)}</td>
                  <td className="py-3 font-medium">{formatCurrency(loan.amount)}</td>
                  <td className="py-3">
                    <Badge variant="info">{loan.status}</Badge>
                  </td>
                  <td className="py-3">{loan.interest_rate}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  );
}
