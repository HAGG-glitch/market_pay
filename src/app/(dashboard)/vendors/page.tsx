"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Plus } from "lucide-react";

const mockVendors = [
  { id: "v_001", name: "Grace Okafor", phone: "+234 801 234 5678", kyc_status: "VERIFIED", credit_score: 78 },
  { id: "v_002", name: "Musa Bello", phone: "+234 802 345 6789", kyc_status: "VERIFIED", credit_score: 82 },
  { id: "v_003", name: "Chioma Eze", phone: "+234 803 456 7890", kyc_status: "PENDING", credit_score: 45 },
  { id: "v_004", name: "Segun Adeleke", phone: "+234 804 567 8901", kyc_status: "VERIFIED", credit_score: 60 },
  { id: "v_005", name: "Amina Yusuf", phone: "+234 805 678 9012", kyc_status: "REJECTED", credit_score: 30 },
];

export default function VendorsPage() {
  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Vendors</h1>
          <p className="text-gray-500">Manage registered vendors</p>
        </div>
        <Link href="/vendors/onboard">
          <Button aria-label="Add a new vendor">
            <Plus size={16} className="mr-1.5" aria-hidden="true" />
            Add Vendor
          </Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Vendors ({mockVendors.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-gray-50 text-left text-gray-500">
                  <th className="pb-3 pt-3 pl-4 font-medium">Name</th>
                  <th className="pb-3 pt-3 font-medium">Phone</th>
                  <th className="pb-3 pt-3 font-medium">KYC Status</th>
                  <th className="pb-3 pt-3 font-medium">Credit Score</th>
                  <th className="pb-3 pt-3 pr-4 font-medium" />
                </tr>
              </thead>
              <tbody>
                {mockVendors.map((v, i) => (
                  <tr key={v.id} className={`border-b last:border-0 transition-colors hover:bg-gray-100 ${i % 2 === 0 ? "bg-white" : "bg-gray-50/50"}`}>
                    <td className="py-3 pl-4 font-medium text-gray-900">{v.name}</td>
                    <td className="py-3 text-gray-600">{v.phone}</td>
                    <td className="py-3">
                      <Badge
                        variant={
                          v.kyc_status === "VERIFIED"
                            ? "success"
                            : v.kyc_status === "PENDING"
                            ? "warning"
                            : "danger"
                        }
                      >
                        {v.kyc_status}
                      </Badge>
                    </td>
                    <td className="py-3">
                      <span
                        className={`font-medium ${
                          v.credit_score >= 70
                            ? "text-green-600"
                            : v.credit_score >= 50
                            ? "text-yellow-600"
                            : "text-red-600"
                        }`}
                      >
                        {v.credit_score}
                      </span>
                    </td>
                    <td className="py-3 pr-4">
                      <Link
                        href={`/vendors/${v.id}`}
                        className="text-primary hover:underline font-medium"
                        aria-label={`View vendor ${v.name}`}
                      >
                        View
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
