"use client";

import { useState, useRef, useEffect } from "react";
import { useLoans, useUpdateLoanStatus } from "@/hooks/use-loans";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { formatCurrency } from "@/lib/utils";
import { disburseLoan, revertDisbursement } from "@/lib/api/loan.service";
import { useAuthStore } from "@/store/auth.store";
import { useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { Plus, MoreVertical, RefreshCw, AlertCircle, CheckCircle, Clock } from "lucide-react";
import { UserRole } from "@/types";

const statusColors: Record<string, "success" | "warning" | "danger" | "info" | "default"> = {
  DRAFT: "default",
  PENDING_REVIEW: "warning",
  UNDER_REVIEW: "warning",
  APPROVED: "info",
  REJECTED: "danger",
  DISBURSEMENT_PENDING: "default",
  DISBURSED: "info",
  ACTIVE: "success",
  CLOSED: "success",
  DEFAULTED: "danger",
};

const sourceLabel: Record<string, string> = {
  USSD: "USSD",
  WEB: "Web",
};

export default function LoansPage() {
  const [statusFilter, setStatusFilter] = useState("PENDING_REVIEW");
  const [sourceFilter, setSourceFilter] = useState("");
  const [actionMsg, setActionMsg] = useState("");
  const user = useAuthStore((s) => s.user);
  const role = user?.role as UserRole | undefined;
  const queryClient = useQueryClient();
  const updateStatus = useUpdateLoanStatus();

  const params: { status?: string } = {};
  if (statusFilter) params.status = statusFilter;
  const { data: loansData } = useLoans(params);
  const loans = (loansData?.data || []).filter(
    (l) => !sourceFilter || l.source === sourceFilter
  );

  const stateTabs = [
    { label: "Pending Review", value: "PENDING_REVIEW" },
    { label: "Approved", value: "APPROVED" },
    { label: "Pending Disburse", value: "DISBURSEMENT_PENDING" },
    { label: "Active", value: "ACTIVE" },
    { label: "All", value: "" },
  ];

  const isDisburser =
    role && [UserRole.LOAN_OFFICER, UserRole.ADMIN, UserRole.SUPER_ADMIN].includes(role);

  const handleRevert = async (loanId: string) => {
    setActionMsg("");
    try {
      await revertDisbursement(loanId);
      queryClient.invalidateQueries({ queryKey: ["loans"] });
      setActionMsg("Disbursement reverted to Approved");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed";
      setActionMsg(msg);
    }
  };

  const handleRetryDisburse = async (loanId: string) => {
    setActionMsg("");
    try {
      await disburseLoan(loanId, "");
      queryClient.invalidateQueries({ queryKey: ["loans"] });
      setActionMsg("Disbursement initiated");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed";
      setActionMsg(msg);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Loans</h1>
          <p className="text-gray-500">Track all loan applications</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => queryClient.invalidateQueries({ queryKey: ["loans"] })}
          >
            <RefreshCw size={14} className="mr-1" />
            Refresh
          </Button>
          <Link href="/loans/apply">
            <Button aria-label="Apply for a new loan">
              <Plus size={16} className="mr-1.5" aria-hidden="true" />
              Apply for Loan
            </Button>
          </Link>
        </div>
      </div>

      {actionMsg && (
        <div className="rounded-lg border border-blue-200 bg-blue-50 px-4 py-2 text-sm text-blue-700">
          {actionMsg}
        </div>
      )}

      <div className="flex flex-wrap gap-2">
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
        <span className="mx-1 w-px bg-gray-300" />
        {[
          { label: "All Sources", value: "" },
          { label: "Web", value: "WEB" },
          { label: "USSD", value: "USSD" },
        ].map((s) => (
          <button
            key={s.value}
            onClick={() => setSourceFilter(s.value)}
            className={`rounded-md px-2 py-1 text-xs font-medium transition-colors ${
              sourceFilter === s.value
                ? "bg-gray-800 text-white"
                : "bg-gray-100 text-gray-500 hover:bg-gray-200"
            }`}
          >
            {s.label}
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
                    <th className="pb-3 pt-3 pl-4 font-medium">Vendor</th>
                    <th className="pb-3 pt-3 font-medium">Amount</th>
                    <th className="pb-3 pt-3 font-medium">Interest</th>
                    <th className="pb-3 pt-3 font-medium">Status</th>
                    <th className="pb-3 pt-3 font-medium">Source</th>
                    <th className="pb-3 pt-3 font-medium">Payout Ref</th>
                    <th className="pb-3 pt-3 pr-4 font-medium" />
                  </tr>
                </thead>
                <tbody>
                  {loans.map((loan, i) => (
                    <Row
                      key={loan.id}
                      loan={loan}
                      i={i}
                      isDisburser={!!isDisburser}
                      onRetry={handleRetryDisburse}
                      onRevert={handleRevert}
                    />
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

function Row({
  loan,
  i,
  isDisburser,
  onRetry,
  onRevert,
}: {
  loan: NonNullable<ReturnType<typeof useLoans>["data"]>["data"][number];
  i: number;
  isDisburser: boolean;
  onRetry: (id: string) => void;
  onRevert: (id: string) => void;
}) {
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const canRetry =
    isDisburser &&
    ((loan.status === "APPROVED" && !loan.monime_reference) || loan.status === "DISBURSED");

  const payoutPending = loan.status === "DISBURSEMENT_PENDING";
  const payoutFailed = (loan.status === "APPROVED" && !loan.monime_reference) || (loan.status === "DISBURSED" && !loan.monime_reference);

  const sourceBadgeColor = loan.source === "USSD" ? "default" : "info";

  return (
    <tr
      className={`border-b last:border-0 transition-colors hover:bg-gray-100 ${
        i % 2 === 0 ? "bg-white" : "bg-gray-50/50"
      }`}
    >
      <td className="py-3 pl-4 text-sm text-gray-900">
        <Link href={`/loans/${loan.id}`} className="hover:text-primary">
          {loan.vendor_name || loan.vendor_id.slice(0, 8) + "..."}
        </Link>
      </td>
      <td className="py-3 font-medium">{formatCurrency(loan.amount)}</td>
      <td className="py-3">{loan.interest_rate}%</td>
      <td className="py-3">
        <Badge variant={statusColors[loan.status] || "default"}>
          {loan.status.replace("_", " ")}
        </Badge>
      </td>
      <td className="py-3">
        <Badge variant={sourceBadgeColor}>
          {sourceLabel[loan.source] || loan.source}
        </Badge>
      </td>
      <td className="py-3 font-mono text-xs text-gray-400">
        {loan.monime_reference ? (
          <span className="flex items-center gap-1">
            <CheckCircle size={12} className="text-green-500" />
            {loan.monime_reference.slice(0, 12)}...
          </span>
        ) : payoutPending ? (
          <span className="flex items-center gap-1 text-yellow-500">
            <Clock size={12} />
            Processing
          </span>
        ) : payoutFailed ? (
          <span className="flex items-center gap-1 text-red-500">
            <AlertCircle size={12} />
            Failed
          </span>
        ) : (
          <span className="flex items-center gap-1 text-gray-300">
            <Clock size={12} />
            Pending
          </span>
        )}
      </td>
      <td className="py-3 pr-4">
        <div className="relative" ref={menuRef}>
          <button
            type="button"
            onClick={() => setMenuOpen(!menuOpen)}
            className="rounded p-1 hover:bg-gray-200"
            aria-label="Actions"
          >
            <MoreVertical size={16} className="text-gray-500" />
          </button>
          {menuOpen && (
            <div className="absolute right-0 z-50 mt-1 w-48 rounded-lg border bg-white py-1 shadow-lg">
              {canRetry && (
                <button
                  type="button"
                  onClick={() => {
                    onRetry(loan.id);
                    setMenuOpen(false);
                  }}
                  className="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                >
                  <RefreshCw size={14} />
                  Retry Disbursement
                </button>
              )}
              {isDisburser && loan.status === "ACTIVE" && loan.monime_reference && (
                <button
                  type="button"
                  onClick={() => {
                    onRevert(loan.id);
                    setMenuOpen(false);
                  }}
                  className="flex w-full items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                >
                  <AlertCircle size={14} />
                  Revert Disbursement
                </button>
              )}
              <Link
                href={`/loans/${loan.id}`}
                className="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                onClick={() => setMenuOpen(false)}
              >
                View Details
              </Link>
            </div>
          )}
        </div>
      </td>
    </tr>
  );
}
