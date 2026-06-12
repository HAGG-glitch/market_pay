"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import Link from "next/link";
import { Plus, Snowflake, RotateCcw, CheckCircle, UserCog } from "lucide-react";
import { getVendors, freezeVendor, unfreezeVendor, approveVendorKYC, assignFieldAgent, getFieldAgents } from "@/lib/api/vendor.service";
import { useAuthStore } from "@/store/auth.store";
import { UserRole } from "@/types";

const canManage: UserRole[] = [UserRole.LOAN_OFFICER, UserRole.ADMIN, UserRole.SUPER_ADMIN];

export default function VendorsPage() {
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const role = user?.role as UserRole | undefined;
  const [freezeModal, setFreezeModal] = useState<{ id: string; name: string } | null>(null);
  const [freezeReason, setFreezeReason] = useState("");
  const [assignModal, setAssignModal] = useState<{ id: string; name: string } | null>(null);
  const [selectedFieldAgent, setSelectedFieldAgent] = useState("");

  const { data: vendors = [], isLoading } = useQuery({
    queryKey: ["vendors"],
    queryFn: getVendors,
  });

  const { data: fieldAgents = [] } = useQuery({
    queryKey: ["field-agents"],
    queryFn: getFieldAgents,
    enabled: !!assignModal,
  });

  const freezeMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => freezeVendor(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vendors"] });
      setFreezeModal(null);
      setFreezeReason("");
    },
  });

  const unfreezeMutation = useMutation({
    mutationFn: (id: string) => unfreezeVendor(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["vendors"] }),
  });

  const approveKYCMutation = useMutation({
    mutationFn: (id: string) => approveVendorKYC(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["vendors"] }),
  });

  const assignMutation = useMutation({
    mutationFn: ({ vendorId, agentId }: { vendorId: string; agentId: string }) => assignFieldAgent(vendorId, agentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vendors"] });
      setAssignModal(null);
      setSelectedFieldAgent("");
    },
  });

  const isFrozen = (v: { kyc_status: string }) =>
    v.kyc_status === "SUSPENDED" || v.kyc_status === "BLACKLISTED";

  const isManage = role ? canManage.includes(role) : false;

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
          <CardTitle>All Vendors ({vendors.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-gray-500">Loading vendors...</p>
          ) : vendors.length === 0 ? (
            <p className="text-sm text-gray-500">No vendors in this mode yet.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-gray-50 text-left text-gray-500">
                    <th className="pb-3 pt-3 pl-4 font-medium">Name</th>
                    <th className="pb-3 pt-3 font-medium">Phone</th>
                    <th className="pb-3 pt-3 font-medium">KYC Status</th>
                    <th className="pb-3 pt-3 font-medium">Credit Score</th>
                    {isManage && <th className="pb-3 pt-3 pr-4 font-medium">Actions</th>}
                  </tr>
                </thead>
                <tbody>
                  {vendors.map((v, i) => (
                    <tr
                      key={v.id}
                      className={`border-b last:border-0 transition-colors hover:bg-gray-100 ${
                        i % 2 === 0 ? "bg-white" : "bg-gray-50/50"
                      }`}
                    >
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
                      <td className="py-3">{v.credit_score}</td>
                      {isManage && (
                        <td className="py-3 pr-4">
                          <div className="flex gap-1">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setAssignModal({ id: v.id, name: v.name })}
                              title="Assign Field Agent"
                              className="h-8 w-8 p-0 text-blue-600"
                            >
                              <UserCog size={16} />
                            </Button>
                            {v.kyc_status === "PENDING" && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => approveKYCMutation.mutate(v.id)}
                                disabled={approveKYCMutation.isPending}
                                title="Approve KYC"
                                className="h-8 w-8 p-0 text-green-600"
                              >
                                <CheckCircle size={16} />
                              </Button>
                            )}
                            {isFrozen(v) ? (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => unfreezeMutation.mutate(v.id)}
                                disabled={unfreezeMutation.isPending}
                                title="Unfreeze"
                                className="h-8 w-8 p-0 text-amber-600"
                              >
                                <RotateCcw size={16} />
                              </Button>
                            ) : (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setFreezeModal({ id: v.id, name: v.name })}
                                title="Freeze"
                                className="h-8 w-8 p-0 text-red-600"
                              >
                                <Snowflake size={16} />
                              </Button>
                            )}
                          </div>
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {freezeModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900">
              Freeze {freezeModal.name}
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              This will suspend the vendor&apos;s account.
            </p>
            <div className="mt-4">
              <Input
                id="freezeReason"
                label="Reason"
                value={freezeReason}
                onChange={(e) => setFreezeReason(e.target.value)}
                placeholder="Enter reason for freezing"
                required
              />
            </div>
            <div className="mt-4 flex gap-3 justify-end">
              <Button
                variant="outline"
                onClick={() => { setFreezeModal(null); setFreezeReason(""); }}
              >
                Cancel
              </Button>
              <Button
                variant="danger"
                onClick={() =>
                  freezeMutation.mutate({ id: freezeModal.id, reason: freezeReason })
                }
                disabled={!freezeReason || freezeMutation.isPending}
              >
                {freezeMutation.isPending ? "Freezing..." : "Confirm Freeze"}
              </Button>
            </div>
          </div>
        </div>
      )}

      {assignModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900">
              Assign Field Agent — {assignModal.name}
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              Select a field agent to assign to this vendor.
            </p>
            <div className="mt-4">
              <label className="text-sm font-medium text-gray-700">Field Agent</label>
              <select
                className="mt-1 flex h-10 w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
                value={selectedFieldAgent}
                onChange={(e) => setSelectedFieldAgent(e.target.value)}
              >
                <option value="">Choose a field agent...</option>
                {fieldAgents.map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.display_name || a.email || a.phone}
                  </option>
                ))}
              </select>
            </div>
            <div className="mt-4 flex gap-3 justify-end">
              <Button
                variant="outline"
                onClick={() => { setAssignModal(null); setSelectedFieldAgent(""); }}
              >
                Cancel
              </Button>
              <Button
                onClick={() =>
                  assignMutation.mutate({ vendorId: assignModal.id, agentId: selectedFieldAgent })
                }
                disabled={!selectedFieldAgent || assignMutation.isPending}
              >
                {assignMutation.isPending ? "Assigning..." : "Assign"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
