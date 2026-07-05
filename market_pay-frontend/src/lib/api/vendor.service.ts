import apiClient from "./client";
import type { PaginatedResponse, Vendor } from "@/types";

export interface FieldAgentUser {
  id: string;
  email: string;
  phone: string;
  display_name: string;
}

interface BackendVendor {
  id: string;
  first_name: string;
  last_name: string;
  phone: string;
  kyc_status: string;
  group_id?: string | null;
  credit_score?: number;
}

function mapVendor(v: BackendVendor): Vendor {
  return {
    id: v.id,
    name: `${v.first_name} ${v.last_name}`.trim(),
    phone: v.phone,
    kyc_status: v.kyc_status,
    group_id: v.group_id ?? null,
    credit_score: v.credit_score ?? 0,
  };
}

export async function getVendors() {
  const { data } = await apiClient.get<PaginatedResponse<BackendVendor>>(
    "/vendors",
    { params: { limit: 100 } }
  );
  return data.data.map(mapVendor);
}

export async function getVendor(id: string) {
  const { data } = await apiClient.get<BackendVendor>(`/vendors/${id}`);
  return mapVendor(data);
}

export async function onboardVendor(payload: {
  name: string;
  phone: string;
  pin: string;
  national_id_number: string;
  national_id_type: string;
  date_of_birth: string;
  market_association_id: string;
  business_name?: string;
  address?: string;
}) {
  const [firstName, ...rest] = payload.name.trim().split(" ");
  const { data } = await apiClient.post<BackendVendor>("/vendors", {
    first_name: firstName || payload.name,
    last_name: rest.join(" ") || "-",
    phone: payload.phone,
    pin: payload.pin,
    national_id_number: payload.national_id_number,
    national_id_type: payload.national_id_type,
    date_of_birth: payload.date_of_birth,
    market_association_id: payload.market_association_id,
    business_name: payload.business_name || "",
    address: payload.address || "",
  });
  return mapVendor(data);
}

export async function freezeVendor(id: string, reason: string) {
  const { data } = await apiClient.put(`/vendors/${id}/freeze`, { reason });
  return data;
}

export async function unfreezeVendor(id: string) {
  const { data } = await apiClient.put(`/vendors/${id}/unfreeze`);
  return data;
}

export async function approveVendorKYC(id: string) {
  const { data } = await apiClient.put(`/vendors/${id}/kyc/approve`);
  return data;
}

export async function assignFieldAgent(vendorId: string, fieldAgentId: string) {
  const { data } = await apiClient.put(`/vendors/${vendorId}/field-agent`, { field_agent_id: fieldAgentId });
  return data;
}

export async function getFieldAgents() {
  const { data } = await apiClient.get<FieldAgentUser[]>("/auth/users?role=FIELD_AGENT");
  return data;
}
