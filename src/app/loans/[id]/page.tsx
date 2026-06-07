"use client";

import { useParams } from "next/navigation";
import { useLoan } from "@/hooks/use-loans";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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

export default function LoanDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const { data: loan, isLoading } = useLoan(id);

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-[#486B6D] border-t-transparent" />
      </div>
    );
  }

  if (!loan) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Loan not found</h1>
        <Link href="/loans">
          <Button variant="outline">Back to Loans</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/loans">
          <Button variant="ghost" size="sm">&larr; Back</Button>
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            Loan {loan.id.slice(0, 8)}
          </h1>
          <p className="text-gray-500">Loan status tracker</p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Amount</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatCurrency(loan.amount)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Interest</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{loan.interest_rate}%</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Status</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant={statusColors[loan.status] || "default"}>
              {loan.status}
            </Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Total Due</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {formatCurrency(
                loan.repayment_schedule.reduce((s, r) => s + r.amount, 0)
              )}
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Loan Status Progress</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            {["DRAFT", "PENDING_REVIEW", "APPROVED", "DISBURSED", "ACTIVE", "CLOSED"].map(
              (s, i) => {
                const statuses = ["DRAFT", "PENDING_REVIEW", "APPROVED", "DISBURSED", "ACTIVE", "CLOSED"];
                const currentIdx = statuses.indexOf(loan.status);
                const isComplete = i <= currentIdx;
                const isCurrent = statuses[i] === loan.status;

                return (
                  <div key={s} className="flex items-center gap-2">
                    <div
                      className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-medium ${
                        isComplete
                          ? "bg-[#486B6D] text-white"
                          : "bg-gray-100 text-gray-400"
                      } ${isCurrent ? "ring-2 ring-[#486B6D] ring-offset-2" : ""}`}
                    >
                      {i + 1}
                    </div>
                    <span className="text-xs text-gray-500">{s.replace("_", " ")}</span>
                    {i < 5 && <div className={`h-px w-6 ${isComplete ? "bg-[#486B6D]" : "bg-gray-200"}`} />}
                  </div>
                );
              }
            )}
          </div>
        </CardContent>
      </Card>

      {loan.repayment_schedule.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Repayment Schedule</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {loan.repayment_schedule.map((s, i) => (
                <div
                  key={i}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className={`flex h-8 w-8 items-center justify-center rounded-full text-xs font-medium ${
                        s.paid
                          ? "bg-green-100 text-green-700"
                          : "bg-gray-100 text-gray-500"
                      }`}
                    >
                      {i + 1}
                    </div>
                    <div>
                      <p className="text-sm font-medium text-gray-900">
                        {formatCurrency(s.amount)}
                      </p>
                      <p className="text-xs text-gray-500">{formatDate(s.due_date)}</p>
                    </div>
                  </div>
                  <Badge variant={s.paid ? "success" : "warning"}>
                    {s.paid ? "Paid" : "Pending"}
                  </Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
