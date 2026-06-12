"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import Link from "next/link";
import { Plus, Snowflake, RotateCcw, UserPlus, ExternalLink } from "lucide-react";
import { getGroups, createGroup, freezeGroup, unfreezeGroup, addGroupMember } from "@/lib/api/group.service";
import { getVendors } from "@/lib/api/vendor.service";
import { useAuthStore } from "@/store/auth.store";
import { UserRole } from "@/types";

const canManage: UserRole[] = [UserRole.LOAN_OFFICER, UserRole.ADMIN, UserRole.SUPER_ADMIN];

export default function GroupLendingPage() {
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const role = user?.role as UserRole | undefined;
  const isManage = role ? canManage.includes(role) : false;

  const [showCreate, setShowCreate] = useState(false);
  const [createName, setCreateName] = useState("");
  const [createDesc, setCreateDesc] = useState("");

  const [freezeModal, setFreezeModal] = useState<{ id: string; name: string } | null>(null);
  const [freezeReason, setFreezeReason] = useState("");

  const [addMemberModal, setAddMemberModal] = useState<{ id: string; name: string } | null>(null);
  const [memberVendorId, setMemberVendorId] = useState("");

  const { data: groups = [], isLoading } = useQuery({
    queryKey: ["groups"],
    queryFn: getGroups,
  });

  const { data: vendors = [] } = useQuery({
    queryKey: ["vendors"],
    queryFn: getVendors,
    enabled: !!addMemberModal,
  });

  const totalMembers = groups.reduce(
    (s, g) => s + (g.members?.length ?? 0),
    0
  );
  const frozen = groups.filter((g) => g.status === "FROZEN").length;

  const createMutation = useMutation({
    mutationFn: () => createGroup(createName, createDesc),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["groups"] });
      setShowCreate(false);
      setCreateName("");
      setCreateDesc("");
    },
  });

  const freezeMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => freezeGroup(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["groups"] });
      setFreezeModal(null);
      setFreezeReason("");
    },
  });

  const unfreezeMutation = useMutation({
    mutationFn: (id: string) => unfreezeGroup(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["groups"] }),
  });

  const addMemberMutation = useMutation({
    mutationFn: () => addGroupMember(addMemberModal!.id, memberVendorId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["groups"] });
      setAddMemberModal(null);
      setMemberVendorId("");
    },
  });

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Group Lending</h1>
          <p className="text-gray-500">Vendor group overview</p>
        </div>
        <Button onClick={() => setShowCreate(true)} aria-label="Create a new group">
          <Plus size={16} className="mr-1.5" aria-hidden="true" />
          Create Group
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Total Groups
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{groups.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Total Members
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-green-600">{totalMembers}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Frozen Groups
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-red-600">{frozen}</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Groups</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-gray-500">Loading groups...</p>
          ) : groups.length === 0 ? (
            <p className="text-sm text-gray-500">No groups in this mode yet.</p>
          ) : (
            <div className="space-y-3">
              {groups.map((g) => (
                <div
                  key={g.id}
                  className="flex items-center justify-between rounded-lg border p-4"
                >
                  <div className="flex-1">
                    <Link
                      href={`/group-lending/${g.id}`}
                      className="font-medium text-gray-900 hover:text-primary hover:underline"
                    >
                      {g.name}
                    </Link>
                    <p className="text-sm text-gray-500">
                      {g.members?.length ?? 0} members
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge
                      variant={g.status === "FROZEN" ? "danger" : "success"}
                    >
                      {g.status}
                    </Badge>
                    <Link href={`/group-lending/${g.id}`}>
                      <Button variant="ghost" size="sm" className="h-8 w-8 p-0" title="View details">
                        <ExternalLink size={16} />
                      </Button>
                    </Link>
                    {isManage && (
                      <>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-blue-600"
                          title="Add member"
                          onClick={() => setAddMemberModal({ id: g.id, name: g.name })}
                        >
                          <UserPlus size={16} />
                        </Button>
                        {g.status === "FROZEN" ? (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0 text-amber-600"
                            title="Unfreeze"
                            onClick={() => unfreezeMutation.mutate(g.id)}
                            disabled={unfreezeMutation.isPending}
                          >
                            <RotateCcw size={16} />
                          </Button>
                        ) : (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0 text-red-600"
                            title="Freeze"
                            onClick={() => setFreezeModal({ id: g.id, name: g.name })}
                          >
                            <Snowflake size={16} />
                          </Button>
                        )}
                      </>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900">Create Group</h3>
            <div className="mt-4 space-y-4">
              <Input
                id="groupName"
                label="Group Name"
                value={createName}
                onChange={(e) => setCreateName(e.target.value)}
                placeholder="Enter group name"
                required
              />
              <Input
                id="groupDesc"
                label="Description"
                value={createDesc}
                onChange={(e) => setCreateDesc(e.target.value)}
                placeholder="Optional description"
              />
            </div>
            <div className="mt-4 flex gap-3 justify-end">
              <Button variant="outline" onClick={() => { setShowCreate(false); setCreateName(""); setCreateDesc(""); }}>
                Cancel
              </Button>
              <Button onClick={() => createMutation.mutate()} disabled={!createName || createMutation.isPending}>
                {createMutation.isPending ? "Creating..." : "Create"}
              </Button>
            </div>
          </div>
        </div>
      )}

      {freezeModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900">
              Freeze {freezeModal.name}
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              This will suspend all group activity.
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
              <Button variant="outline" onClick={() => { setFreezeModal(null); setFreezeReason(""); }}>
                Cancel
              </Button>
              <Button
                variant="danger"
                onClick={() => freezeMutation.mutate({ id: freezeModal.id, reason: freezeReason })}
                disabled={!freezeReason || freezeMutation.isPending}
              >
                {freezeMutation.isPending ? "Freezing..." : "Confirm Freeze"}
              </Button>
            </div>
          </div>
        </div>
      )}

      {addMemberModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900">
              Add Member to {addMemberModal.name}
            </h3>
            <div className="mt-4">
              <label className="text-sm font-medium text-gray-700">Select Vendor</label>
              <select
                className="mt-1 flex h-10 w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
                value={memberVendorId}
                onChange={(e) => setMemberVendorId(e.target.value)}
              >
                <option value="">Choose a vendor...</option>
                {vendors.map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.name} ({v.phone})
                  </option>
                ))}
              </select>
            </div>
            <div className="mt-4 flex gap-3 justify-end">
              <Button variant="outline" onClick={() => { setAddMemberModal(null); setMemberVendorId(""); }}>
                Cancel
              </Button>
              <Button
                onClick={() => addMemberMutation.mutate()}
                disabled={!memberVendorId || addMemberMutation.isPending}
              >
                {addMemberMutation.isPending ? "Adding..." : "Add Member"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
