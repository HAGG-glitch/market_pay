"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useApplyLoan } from "@/hooks/use-loans";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCurrency } from "@/lib/utils";

export default function LoanApplyPage() {
  const router = useRouter();
  const applyLoan = useApplyLoan();

  const [step, setStep] = useState(1);
  const [amount, setAmount] = useState("");
  const [interestRate, setInterestRate] = useState("5");
  const [tenure, setTenure] = useState("3");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const schedule = Array.from({ length: Number(tenure) }, (_, i) => {
      const d = new Date();
      d.setMonth(d.getMonth() + i + 1);
      return {
        due_date: d.toISOString().split("T")[0],
        amount: Number(amount) / Number(tenure),
      };
    });

    try {
      await applyLoan.mutateAsync({
        amount: Number(amount),
        interest_rate: Number(interestRate),
        repayment_schedule: schedule,
      });
      router.push("/loans");
    } catch {
      // error handled by mutation
    }
  };

  return (
    <div className="mx-auto max-w-lg space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Apply for Loan</h1>
        <p className="text-gray-500">Complete the steps below</p>
      </div>

      <div className="flex items-center gap-2">
        {[1, 2, 3].map((s) => (
          <div key={s} className="flex items-center gap-2">
            <div
              className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium ${
                step >= s
                  ? "bg-primary text-white"
                  : "bg-gray-100 text-gray-400"
              }`}
            >
              {s}
            </div>
            <span className="text-xs text-gray-500">
              {s === 1 ? "Amount" : s === 2 ? "Terms" : "Confirm"}
            </span>
            {s < 3 && <div className="h-px w-8 bg-gray-200" />}
          </div>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>
            {step === 1
              ? "Select Amount & Type"
              : step === 2
              ? "Repayment Schedule"
              : "Confirmation"}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {step === 1 && (
              <>
                <Input
                  id="amount"
                  label="Loan Amount (LE)"
                  type="number"
                  placeholder="10"
                  min={1}
                  max={50}
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                />
                {amount && (Number(amount) < 1 || Number(amount) > 50) && (
                  <p className="text-sm text-red-500">Amount must be between 1 and 50 SLE</p>
                )}
                <Select
                  id="loan-type"
                  label="Loan Type"
                  value="personal"
                  options={[
                    { value: "personal", label: "Personal Loan" },
                    { value: "business", label: "Business Loan" },
                    { value: "group", label: "Group Loan" },
                  ]}
                />
                <Button
                  type="button"
                  className="w-full"
                  onClick={() => setStep(2)}
                  disabled={!amount || Number(amount) < 1 || Number(amount) > 50}
                >
                  Next
                </Button>
              </>
            )}

            {step === 2 && (
              <>
                <Input
                  id="interest"
                  label="Interest Rate (%)"
                  type="number"
                  value={interestRate}
                  onChange={(e) => setInterestRate(e.target.value)}
                  required
                />
                <Select
                  id="tenure"
                  label="Repayment Period"
                  value={tenure}
                  onChange={(e) => setTenure(e.target.value)}
                  options={[
                    { value: "1", label: "1 Month" },
                    { value: "3", label: "3 Months" },
                    { value: "6", label: "6 Months" },
                    { value: "12", label: "12 Months" },
                  ]}
                />
                <div className="flex gap-3">
                  <Button
                    type="button"
                    variant="outline"
                    className="flex-1"
                    onClick={() => setStep(1)}
                  >
                    Back
                  </Button>
                  <Button
                    type="button"
                    className="flex-1"
                    onClick={() => setStep(3)}
                  >
                    Next
                  </Button>
                </div>
              </>
            )}

            {step === 3 && (
              <>
                <div className="space-y-3 rounded-lg bg-gray-50 p-4">
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Amount</span>
                    <span className="font-medium">{formatCurrency(Number(amount))}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Interest Rate</span>
                    <span className="font-medium">{interestRate}%</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Tenure</span>
                    <span className="font-medium">{tenure} months</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Monthly Payment</span>
                    <span className="font-medium">
                      {formatCurrency(Number(amount) / Number(tenure))}
                    </span>
                  </div>
                </div>

                <div className="flex gap-3">
                  <Button
                    type="button"
                    variant="outline"
                    className="flex-1"
                    onClick={() => setStep(2)}
                  >
                    Back
                  </Button>
                  <Button
                    type="submit"
                    className="flex-1"
                    disabled={applyLoan.isPending}
                  >
                    {applyLoan.isPending ? "Submitting..." : "Confirm & Submit"}
                  </Button>
                </div>

                {applyLoan.isError && (
                  <p className="text-sm text-red-500">
                    Failed to submit application. Please try again.
                  </p>
                )}
              </>
            )}
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
