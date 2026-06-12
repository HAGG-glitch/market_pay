"use client";

import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getVendors } from "@/lib/api/vendor.service";
import { getGroups } from "@/lib/api/group.service";
import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function FieldAgentDashboard() {
  const { data: vendors = [] } = useQuery({
    queryKey: ["vendors"],
    queryFn: getVendors,
  });
  const { data: groups = [] } = useQuery({
    queryKey: ["groups"],
    queryFn: getGroups,
  });

  const frozenVendors = vendors.filter((v) => v.kyc_status === "SUSPENDED").length;
  const pendingKyc = vendors.filter((v) => v.kyc_status === "PENDING").length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Field Agent Dashboard</h1>
          <p className="text-gray-500">My vendors and groups</p>
        </div>
        <Link href="/vendors/onboard">
          <Button>Register Vendor</Button>
        </Link>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">My Vendors</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{vendors.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">My Groups</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{groups.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Pending KYC</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-amber-600">{pendingKyc}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Frozen</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-red-600">{frozenVendors}</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
