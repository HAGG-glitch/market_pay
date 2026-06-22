"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import { useLoan, useUpdateLoanStatus } from "@/hooks/use-loans";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { formatCurrency, formatDate } from "@/lib/utils";
import { disburseLoan } from "@/lib/api/loan.service";
import { useAuthStore } from "@/store/auth.store";
import { useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { UserRole } from "@/types";

const statusColors: Record<string, "success" | "warning" | "danger" | "info" | "default"> = {
  DRAFT: "default",
  PENDING_REVIEW: "warning",
  APPROVED: "info",
  DISBURSED: "info",
  ACTIVE: "success",
  CLOSED: "success",
  DEFAULTED: "danger",
};

const canReview: UserRole[] = [UserRole.LOAN_OFFICER, UserRole.ADMIN, UserRole.SUPER_ADMIN];
const canDisburse: UserRole[] = [UserRole.ADMIN, UserRole.SUPER_ADMIN];

export default function LoanDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const user = useAuthStore((s) => s.user);
  const role = user?.role as UserRole | undefined;
  const queryClient = useQueryClient();
  const { data: loan, isLoading } = useLoan(id);
  const updateStatus = useUpdateLoanStatus();

  const [disburseModal, setDisburseModal] = useState(false);
  const [monimeRef, setMonimeRef] = useState("");
  const [actionError, setActionError] = useState("");

  const isReviewer = role ? canReview.includes(role) : false;
  const isDisburser = role ? canDisburse.includes(role) : false;

  const handleAction = (status: string) => {
    setActionError("");
    updateStatus.mutate(
      { id, status },
      {
        onError: (err: Error) => setActionError(err.message),
      }
    );
  };

  const handleDisburse = async () => {
    setActionError("");
    try {
      await disburseLoan(id, monimeRef);
      queryClient.invalidateQueries({ queryKey: ["loan", id] });
      queryClient.invalidateQueries({ queryKey: ["loans"] });
      setDisburseModal(false);
      setMonimeRef("");
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Disbursement failed";
      setActionError(message);
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!loan) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Loan not found</h1>
        <Link href="/loans">
          <Button variant="outline">&larr; Back to Loans</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center gap-3">
        <Link href="/loans">
          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
            <ArrowLeft size={18} />
          </Button>
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            Loan {loan.id.slice(0, 8)}
          </h1>
          <p className="text-gray-500">Loan status tracker</p>
        </div>
      </div>

      {actionError && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {actionError}
        </div>
      )}

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

      {isReviewer && (loan.status === "PENDING_REVIEW" || loan.status === "UNDER_REVIEW") && (
        <Card>
          <CardHeader>
            <CardTitle>Actions</CardTitle>
          </CardHeader>
          <CardContent className="flex gap-3">
            <Button
              variant="default"
              onClick={() => handleAction("APPROVED")}
              disabled={updateStatus.isPending}
            >
              {updateStatus.isPending ? "Processing..." : "Approve"}
            </Button>
            <Button
              variant="danger"
              onClick={() => handleAction("REJECTED")}
              disabled={updateStatus.isPending}
            >
              {updateStatus.isPending ? "Processing..." : "Reject"}
            </Button>
          </CardContent>
        </Card>
      )}

      {isDisburser && loan.status === "APPROVED" && (
        <Card>
          <CardHeader>
            <CardTitle>Disbursement</CardTitle>
          </CardHeader>
          <CardContent>
            <Button onClick={() => setDisburseModal(true)}>
              Disburse Loan
            </Button>
          </CardContent>
        </Card>
      )}

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
                          ? "bg-primary text-white"
                          : "bg-gray-100 text-gray-400"
                      } ${isCurrent ? "ring-2 ring-primary ring-offset-2" : ""}`}
                    >
                      {i + 1}
                    </div>
                    <span className="text-xs text-gray-500">{s.replace("_", " ")}</span>
                    {i < 5 && <div className={`h-px w-6 ${isComplete ? "bg-primary" : "bg-gray-200"}`} />}
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

      {disburseModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900">Disburse Loan</h3>
            <p className="mt-1 text-sm text-gray-500">
              Enter the Monime reference for this disbursement.
            </p>
            <div className="mt-4">
              <Input
                id="monimeRef"
                label="Monime Reference"
                value={monimeRef}
                onChange={(e) => setMonimeRef(e.target.value)}
                placeholder="Enter Monime transaction reference"
                required
              />
            </div>
            <div className="mt-4 flex gap-3 justify-end">
              <Button variant="outline" onClick={() => { setDisburseModal(false); setMonimeRef(""); }}>
                Cancel
              </Button>
              <Button
                onClick={handleDisburse}
                disabled={!monimeRef}
              >
                Confirm Disburse
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
