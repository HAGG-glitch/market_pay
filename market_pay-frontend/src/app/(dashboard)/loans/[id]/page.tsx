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
import { ArrowLeft, CheckCircle, XCircle, Clock, RefreshCw } from "lucide-react";
import { UserRole } from "@/types";

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

const canReview: UserRole[] = [UserRole.LOAN_OFFICER, UserRole.ADMIN, UserRole.SUPER_ADMIN];
const canDisburse: UserRole[] = [UserRole.LOAN_OFFICER, UserRole.ADMIN, UserRole.SUPER_ADMIN];

const timelineStages = [
  { key: "PENDING_REVIEW", label: "Requested", icon: Clock },
  { key: "APPROVED", label: "Approved", icon: CheckCircle },
  { key: "DISBURSED", label: "Disbursement", icon: RefreshCw },
  { key: "ACTIVE", label: "Active", icon: CheckCircle },
  { key: "CLOSED", label: "Completed", icon: CheckCircle },
];

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

  const currentIdx = timelineStages.findIndex((s) => s.key === loan.status);
  const statusOrder = timelineStages.map((s) => s.key);
  const isRejected = loan.status === "REJECTED";

  const stageTimestamps: Record<string, string | undefined> = {
    PENDING_REVIEW: loan.created_at,
    DISBURSED: loan.disbursed_at,
  };

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
              {loan.status.replace("_", " ")}
            </Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Source</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant={loan.source === "USSD" ? "default" : "info"}>
              {loan.source || "WEB"}
            </Badge>
          </CardContent>
        </Card>
      </div>

      {isReviewer && !isRejected && (loan.status === "PENDING_REVIEW" || loan.status === "UNDER_REVIEW") && (
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

      {isDisburser && (loan.status === "APPROVED" || loan.status === "DISBURSED") && (
        <Card>
          <CardHeader>
            <CardTitle>Disbursement</CardTitle>
          </CardHeader>
          <CardContent>
            <Button onClick={() => setDisburseModal(true)}>
              {loan.status === "DISBURSED" ? "Retry Disburse" : "Disburse Loan"}
            </Button>
          </CardContent>
        </Card>
      )}

      {loan.monime_reference && (
        <Card>
          <CardHeader>
            <CardTitle>Payout Details</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-500">Monime Reference</span>
                <span className="font-mono text-xs">{loan.monime_reference}</span>
              </div>
              {loan.disbursed_at && (
                <div className="flex justify-between">
                  <span className="text-gray-500">Disbursed At</span>
                  <span>{formatDate(loan.disbursed_at)}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-gray-500">Status</span>
                <span className="flex items-center gap-1">
                  {loan.status === "ACTIVE" ? (
                    <><CheckCircle size={14} className="text-green-500" /> Successful</>
                  ) : loan.status === "DISBURSED" ? (
                    <><RefreshCw size={14} className="text-yellow-500" /> Pending</>
                  ) : (
                    <><Clock size={14} className="text-gray-400" /> Unknown</>
                  )}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Loan Details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 text-sm md:grid-cols-2">
            <div className="flex justify-between">
              <span className="text-gray-500">Loan ID</span>
              <span className="font-mono text-xs">{loan.id}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Vendor ID</span>
              <span className="font-mono text-xs">{loan.vendor_id}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Applied</span>
              <span>{formatDate(loan.created_at)}</span>
            </div>
            {loan.funded_by && (
              <div className="flex justify-between">
                <span className="text-gray-500">Funded By</span>
                <span>{loan.funded_by}</span>
              </div>
            )}
            {loan.reviewed_by && (
              <div className="flex justify-between">
                <span className="text-gray-500">Reviewed By</span>
                <span className="font-mono text-xs">{loan.reviewed_by}</span>
              </div>
            )}
            {loan.review_note && (
              <div className="flex justify-between md:col-span-2">
                <span className="text-gray-500">Review Note</span>
                <span className="text-gray-700">{loan.review_note}</span>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Loan Timeline</CardTitle>
        </CardHeader>
        <CardContent>
          {isRejected ? (
            <div className="rounded-lg border border-red-200 bg-red-50 p-4">
              <div className="flex items-center gap-2">
                <XCircle size={20} className="text-red-500" />
                <div>
                  <p className="font-medium text-red-800">Loan Rejected</p>
                  {loan.rejection_reason && (
                    <p className="text-sm text-red-600">{loan.rejection_reason}</p>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="space-y-3">
                {timelineStages.map((stage, i) => {
                  const stageIdx = statusOrder.indexOf(stage.key);
                  const isComplete = currentIdx >= stageIdx;
                  const isCurrent = currentIdx === stageIdx;
                  const timestamp = stageTimestamps[stage.key];
                  const Icon = stage.icon;

                  return (
                    <div key={stage.key} className="flex items-start gap-3">
                      <div className="flex flex-col items-center">
                        <div
                          className={`flex h-8 w-8 items-center justify-center rounded-full ${
                            isComplete
                              ? "bg-primary text-white"
                              : "bg-gray-100 text-gray-400"
                          } ${isCurrent ? "ring-2 ring-primary ring-offset-2" : ""}`}
                        >
                          <Icon size={14} />
                        </div>
                        {i < timelineStages.length - 1 && (
                          <div
                            className={`mt-1 h-6 w-0.5 ${
                              isComplete && !isCurrent ? "bg-primary" : "bg-gray-200"
                            }`}
                          />
                        )}
                      </div>
                      <div className="pt-1">
                        <p
                          className={`text-sm font-medium ${
                            isComplete ? "text-gray-900" : "text-gray-400"
                          }`}
                        >
                          {stage.label}
                        </p>
                        {timestamp && (
                          <p className="text-xs text-gray-400">
                            {formatDate(timestamp)}
                          </p>
                        )}
                        {isCurrent && !timestamp && (
                          <p className="text-xs text-gray-500">Current stage</p>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
          )}
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
              Leave blank to auto-disburse via Monime Payout, or enter a reference manually.
            </p>
            <div className="mt-4">
              <Input
                id="monimeRef"
                label="Monime Reference (optional)"
                value={monimeRef}
                onChange={(e) => setMonimeRef(e.target.value)}
                placeholder="Auto-generated if left blank"
              />
            </div>
            <div className="mt-4 flex gap-3 justify-end">
              <Button variant="outline" onClick={() => { setDisburseModal(false); setMonimeRef(""); }}>
                Cancel
              </Button>
              <Button onClick={handleDisburse}>
                {monimeRef ? "Disburse with Reference" : "Auto-Disburse"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
