"use client";

import { useState } from "react";
import { usePayments, useMakePayment } from "@/hooks/use-payments";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { formatCurrency, formatDate } from "@/lib/utils";

export default function PaymentsPage() {
  const { data: paymentsData } = usePayments();
  const makePayment = useMakePayment();
  const payments = paymentsData?.data || [];

  const [vendorId, setVendorId] = useState("");
  const [amount, setAmount] = useState("");

  const handlePayment = (e: React.FormEvent) => {
    e.preventDefault();
    makePayment.mutate(
      { vendor_id: vendorId, amount: Number(amount) },
      {
        onSuccess: () => {
          setVendorId("");
          setAmount("");
        },
      }
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Payments</h1>
        <p className="text-gray-500">Make payments and view history</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Make a Payment</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handlePayment} className="space-y-4">
            <Input
              id="vendorId"
              label="Vendor ID"
              value={vendorId}
              onChange={(e) => setVendorId(e.target.value)}
              placeholder="Enter vendor ID"
              required
            />
            <Input
              id="amount"
              label="Amount (LE)"
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="0.00"
              required
            />
            <Button
              type="submit"
              className="w-full"
              size="lg"
              disabled={makePayment.isPending}
              aria-label="Submit payment"
            >
              {makePayment.isPending ? "Processing..." : "Pay Now"}
            </Button>
            {makePayment.isError && (
              <p className="text-sm text-red-500">
                Payment failed. Please try again.
              </p>
            )}
            {makePayment.isSuccess && (
              <p className="text-sm text-green-600">Payment successful!</p>
            )}
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Payment History</CardTitle>
        </CardHeader>
        <CardContent>
          {payments.length === 0 ? (
            <p className="text-sm text-gray-500">No payments yet.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-gray-50 text-left text-gray-500">
                    <th className="pb-3 pt-3 pl-4 font-medium">ID</th>
                    <th className="pb-3 pt-3 font-medium">Vendor</th>
                    <th className="pb-3 pt-3 font-medium">Amount</th>
                    <th className="pb-3 pt-3 font-medium">Fee</th>
                    <th className="pb-3 pt-3 font-medium">Status</th>
                    <th className="pb-3 pt-3 pr-4 font-medium">Date</th>
                  </tr>
                </thead>
                <tbody>
                  {payments.map((p, i) => (
                    <tr key={p.id} className={`border-b last:border-0 transition-colors hover:bg-gray-100 ${i % 2 === 0 ? "bg-white" : "bg-gray-50/50"}`}>
                      <td className="py-3 pl-4 font-mono text-xs">
                        {p.id.slice(0, 8)}...
                      </td>
                      <td className="py-3">{p.vendor_id.slice(0, 8)}...</td>
                      <td className="py-3 font-medium">
                        {formatCurrency(p.amount)}
                      </td>
                      <td className="py-3">{formatCurrency(p.fee)}</td>
                      <td className="py-3">
                        <span
                          className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                            p.status === "SUCCESS"
                              ? "bg-green-100 text-green-700"
                              : p.status === "FAILED"
                              ? "bg-red-100 text-red-700"
                              : "bg-yellow-100 text-yellow-700"
                          }`}
                        >
                          {p.status}
                        </span>
                      </td>
                      <td className="py-3 pr-4 text-gray-500">
                        {formatDate(p.created_at)}
                      </td>
                    </tr>
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
