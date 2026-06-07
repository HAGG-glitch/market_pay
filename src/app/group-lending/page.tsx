"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

const mockMembers = [
  { id: "1", name: "Grace Okafor", status: "ACTIVE", credit_score: 78 },
  { id: "2", name: "Musa Bello", status: "ACTIVE", credit_score: 82 },
  { id: "3", name: "Chioma Eze", status: "DEFAULTED", credit_score: 45 },
  { id: "4", name: "Segun Adeleke", status: "PENDING", credit_score: 60 },
];

export default function GroupLendingPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Group Lending</h1>
        <p className="text-gray-500">Vendor group overview</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Total Members
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{mockMembers.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Active Members
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-green-600">
              {mockMembers.filter((m) => m.status === "ACTIVE").length}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Defaulted
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-red-600">
              {mockMembers.filter((m) => m.status === "DEFAULTED").length}
            </p>
          </CardContent>
        </Card>
      </div>

      {mockMembers.some((m) => m.status === "DEFAULTED") && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-full bg-red-500" />
              <p className="text-sm font-medium text-red-800">
                Group is frozen due to defaulting member(s)
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Members</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {mockMembers.map((member) => (
              <div
                key={member.id}
                className="flex items-center justify-between rounded-lg border p-4"
              >
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#486B6D]/10 text-sm font-medium text-[#486B6D]">
                    {member.name.charAt(0)}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">
                      {member.name}
                    </p>
                    <p className="text-xs text-gray-500">
                      Credit Score: {member.credit_score}
                    </p>
                  </div>
                </div>
                <Badge
                  variant={
                    member.status === "ACTIVE"
                      ? "success"
                      : member.status === "DEFAULTED"
                      ? "danger"
                      : "warning"
                  }
                >
                  {member.status}
                </Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Button variant="outline" className="w-full">
        Add Member to Group
      </Button>
    </div>
  );
}
