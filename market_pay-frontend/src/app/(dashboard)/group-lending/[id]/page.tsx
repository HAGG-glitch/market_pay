"use client";

import { useParams, useRouter } from "next/navigation";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getGroup, addGroupMember, freezeGroup, unfreezeGroup } from "@/lib/api/group.service";
import { getVendors } from "@/lib/api/vendor.service";
import { useAuthStore } from "@/store/auth.store";
import { useState } from "react";
import { ArrowLeft, UserPlus, Snowflake, RotateCcw } from "lucide-react";
import Link from "next/link";
import { UserRole } from "@/types";

const canManage: UserRole[] = [UserRole.LOAN_OFFICER, UserRole.ADMIN, UserRole.SUPER_ADMIN];

export default function GroupDetailPage() {
  const params = useParams();
  const router = useRouter();
  const queryClient = useQueryClient();
  const id = params.id as string;
  const user = useAuthStore((s) => s.user);
  const role = user?.role as UserRole | undefined;
  const isManage = role ? canManage.includes(role) : false;

  const [showAddMember, setShowAddMember] = useState(false);
  const [memberVendorId, setMemberVendorId] = useState("");

  const { data: group, isLoading } = useQuery({
    queryKey: ["group", id],
    queryFn: () => getGroup(id),
    enabled: !!id,
  });

  const { data: vendors = [] } = useQuery({
    queryKey: ["vendors"],
    queryFn: getVendors,
    enabled: showAddMember,
  });

  const addMemberMutation = useMutation({
    mutationFn: () => addGroupMember(id, memberVendorId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group", id] });
      queryClient.invalidateQueries({ queryKey: ["groups"] });
      setShowAddMember(false);
      setMemberVendorId("");
    },
  });

  const freezeMutation = useMutation({
    mutationFn: (reason: string) => freezeGroup(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group", id] });
      queryClient.invalidateQueries({ queryKey: ["groups"] });
    },
  });

  const unfreezeMutation = useMutation({
    mutationFn: () => unfreezeGroup(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group", id] });
      queryClient.invalidateQueries({ queryKey: ["groups"] });
    },
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!group) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Group not found</h1>
        <Link href="/group-lending">
          <Button variant="outline">&larr; Back to Groups</Button>
        </Link>
      </div>
    );
  }

  const members = group.members ?? [];

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center gap-3">
        <Link href="/group-lending">
          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
            <ArrowLeft size={18} />
          </Button>
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{group.name}</h1>
          <p className="text-gray-500">{group.description || "Lending group"}</p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Status</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant={group.status === "FROZEN" ? "danger" : "success"}>
              {group.status}
            </Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Members</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{members.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Leader</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">
              {members.find((m) => m.is_leader)?.vendor_id?.slice(0, 8) || "—"}
            </p>
          </CardContent>
        </Card>
      </div>

      {isManage && (
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowAddMember(true)}
          >
            <UserPlus size={16} className="mr-1.5" />
            Add Member
          </Button>
          {group.status === "FROZEN" ? (
            <Button
              variant="outline"
              size="sm"
              onClick={() => unfreezeMutation.mutate()}
              disabled={unfreezeMutation.isPending}
              className="text-amber-600"
            >
              <RotateCcw size={16} className="mr-1.5" />
              Unfreeze
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                const reason = window.prompt("Reason for freezing:");
                if (reason) freezeMutation.mutate(reason);
              }}
              disabled={freezeMutation.isPending}
              className="text-red-600"
            >
              <Snowflake size={16} className="mr-1.5" />
              Freeze
            </Button>
          )}
        </div>
      )}

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Members ({members.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {members.length === 0 ? (
            <p className="text-sm text-gray-500">No members yet.</p>
          ) : (
            <div className="space-y-2">
              {members.map((m) => (
                <div
                  key={m.vendor_id}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex items-center gap-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                      {m.vendor_id.slice(0, 2)}
                    </div>
                    <div>
                      <p className="text-sm font-medium text-gray-900">
                        {m.vendor_id.slice(0, 8)}...
                      </p>
                      <p className="text-xs text-gray-500">
                        {m.is_leader ? "Group Leader" : "Member"}
                      </p>
                    </div>
                  </div>
                  {m.is_leader && (
                    <Badge variant="info">Leader</Badge>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {showAddMember && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900">Add Member</h3>
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
              <Button variant="outline" onClick={() => { setShowAddMember(false); setMemberVendorId(""); }}>
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
