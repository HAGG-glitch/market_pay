import apiClient from "./client";
import type { PaginatedResponse } from "@/types";

export interface Group {
  id: string;
  name: string;
  description: string;
  status: string;
  members?: { vendor_id: string; is_leader: boolean }[];
}

export async function getGroups() {
  const { data } = await apiClient.get<PaginatedResponse<Group>>("/groups");
  return data.data ?? [];
}

export async function createGroup(name: string, description = "") {
  const { data } = await apiClient.post<Group>("/groups", { name, description });
  return data;
}

export async function getGroup(id: string) {
  const { data } = await apiClient.get<Group>(`/groups/${id}`);
  return data;
}

export async function addGroupMember(groupId: string, vendorId: string) {
  const { data } = await apiClient.post(`/groups/${groupId}/members`, { vendor_id: vendorId });
  return data;
}

export async function freezeGroup(id: string, reason: string) {
  const { data } = await apiClient.put(`/groups/${id}/freeze`, { reason });
  return data;
}

export async function unfreezeGroup(id: string) {
  const { data } = await apiClient.put(`/groups/${id}/unfreeze`);
  return data;
}
