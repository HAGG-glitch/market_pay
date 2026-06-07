"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useLoans, useUpdateLoanStatus } from "@/hooks/use-loans";
import { formatCurrency, formatDate } from "@/lib/utils";

export default function LoanOfficerDashboard() {
  const { data: loansData } = useLoans({ status: "PENDING_REVIEW" });
  const updateStatus = useUpdateLoanStatus();
  const loans = loansData?.data || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Loan Officer Dashboard</h1>
        <p className="text-gray-500">Review and approve loan applications</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Loan Approval Queue ({loans.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {loans.length === 0 ? (
            <p className="text-sm text-gray-500">No pending applications.</p>
          ) : (
            <div className="space-y-4">
              {loans.map((loan) => (
                <div key={loan.id} className="rounded-lg border p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium text-gray-900">
                        {formatCurrency(loan.amount)}
                      </p>
                      <p className="text-sm text-gray-500">
                        Vendor: {loan.vendor_id.slice(0, 8)} | Interest:{" "}
                        {loan.interest_rate}%
                      </p>
                      <p className="text-xs text-gray-400">
                        {loan.repayment_schedule.length} installments
                      </p>
                    </div>
                    <Badge variant="warning">PENDING_REVIEW</Badge>
                  </div>
                  <div className="mt-3 flex gap-2">
                    <Button
                      size="sm"
                      variant="default"
                      onClick={() =>
                        updateStatus.mutate({ id: loan.id, status: "APPROVED" })
                      }
                      disabled={updateStatus.isPending}
                    >
                      Approve
                    </Button>
                    <Button size="sm" variant="outline">
                      Review Details
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
