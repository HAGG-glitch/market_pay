"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatCurrency } from "@/lib/utils";
import { getPaymentPlans } from "@/lib/api/loan.service";
import { RefreshCw, ChevronDown, ChevronUp } from "lucide-react";

const statusColors: Record<string, "success" | "warning" | "danger" | "info" | "default"> = {
  APPROVED: "info",
  DISBURSEMENT_PENDING: "default",
  ACTIVE: "success",
  CLOSED: "success",
  DEFAULTED: "danger",
};

function ProgressBar({ pct }: { pct: number }) {
  const color =
    pct >= 100 ? "bg-green-500" : pct >= 50 ? "bg-amber-500" : "bg-blue-500";
  return (
    <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
      <div
        className={`h-full rounded-full transition-all duration-500 ${color}`}
        style={{ width: `${Math.min(pct, 100)}%` }}
      />
    </div>
  );
}

export default function PaymentPlansPage() {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [stateFilter, setStateFilter] = useState("");

  const { data, isLoading, refetch } = useQuery({
    queryKey: ["payment-plans", stateFilter],
    queryFn: () => getPaymentPlans({ state: stateFilter || undefined }),
    refetchInterval: 30_000,
  });

  const plans = data?.data ?? [];

  const toggleExpand = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const stateTabs = [
    { label: "All", value: "" },
    { label: "Active", value: "ACTIVE" },
    { label: "Approved", value: "APPROVED" },
    { label: "Closed", value: "CLOSED" },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Payment Plans</h1>
          <p className="text-gray-500">Repayment progress for all loans</p>
        </div>
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw size={14} className="mr-1" />
          Refresh
        </Button>
      </div>

      <div className="flex gap-2 overflow-x-auto pb-2">
        {stateTabs.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setStateFilter(tab.value)}
            className={`whitespace-nowrap rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
              stateFilter === tab.value
                ? "bg-primary text-white"
                : "bg-gray-100 text-gray-600 hover:bg-gray-200"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {isLoading ? (
        <p className="text-sm text-gray-500">Loading payment plans...</p>
      ) : plans.length === 0 ? (
        <p className="text-sm text-gray-500">No loans found.</p>
      ) : (
        <div className="space-y-3">
          {plans.map((plan) => (
            <Card key={plan.id}>
              <CardContent className="p-4">
                <div className="flex flex-col gap-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/10 text-sm font-medium text-primary shrink-0">
                        {plan.vendor_name?.charAt(0) || "?"}
                      </div>
                      <div className="min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">
                          {plan.vendor_name || "Unknown Vendor"}
                        </p>
                        <p className="text-xs text-gray-500">
                          {formatCurrency(plan.total_amount)} total &middot; {plan.schedule_count} installment{(plan.schedule_count || 0) !== 1 ? "s" : ""}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant={statusColors[plan.state] || "default"}>
                        {plan.state}
                      </Badge>
                      <button
                        onClick={() => toggleExpand(plan.id)}
                        className="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                      >
                        {expanded.has(plan.id) ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
                      </button>
                    </div>
                  </div>

                  <div className="flex items-center gap-4 text-sm">
                    <div className="flex-1">
                      <div className="flex justify-between mb-1">
                        <span className="text-gray-500">Paid {formatCurrency(plan.total_paid)}</span>
                        <span className="text-gray-500">{plan.remaining > 0 ? `${formatCurrency(plan.remaining)} left` : "Paid off"}</span>
                      </div>
                      <ProgressBar pct={plan.progress_percent} />
                      <div className="flex justify-between mt-1">
                        <span className="text-xs text-gray-400">{plan.paid_count} of {plan.schedule_count} installments</span>
                        <span className="text-xs font-medium text-gray-600">{Math.round(plan.progress_percent)}%</span>
                      </div>
                    </div>
                  </div>

                  {plan.next_due_date && (
                    <p className="text-xs text-amber-600">
                      Next due: {new Date(plan.next_due_date).toLocaleDateString()}
                    </p>
                  )}

                  {expanded.has(plan.id) && plan.schedules && (
                    <div className="border-t pt-3 space-y-2">
                      {plan.schedules.map((sch) => (
                        <div
                          key={sch.id}
                          className="flex items-center justify-between rounded-lg bg-gray-50 px-3 py-2 text-sm"
                        >
                          <div className="flex items-center gap-3">
                            <span className="font-medium text-gray-700">
                              #{sch.installment_no}
                            </span>
                            <span className="text-gray-500">
                              Due {new Date(sch.due_date).toLocaleDateString()}
                            </span>
                          </div>
                          <div className="flex items-center gap-4">
                            <span className="text-gray-700">
                              {formatCurrency(sch.total_due)}
                            </span>
                            {sch.status === "PAID" ? (
                              <Badge variant="success">Paid</Badge>
                            ) : (
                              <Badge variant="warning">
                                {formatCurrency(sch.total_due - sch.amount_paid)} left
                              </Badge>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
