import apiClient from "./client";
import type { Vendor } from "@/types";

export async function getVendors() {
  const { data } = await apiClient.get<Vendor[]>("/vendors");
  return data;
}

export async function getVendor(id: string) {
  const { data } = await apiClient.get<Vendor>(`/vendors/${id}`);
  return data;
}

export async function onboardVendor(payload: {
  name: string;
  phone: string;
}) {
  const { data } = await apiClient.post<Vendor>("/vendors", payload);
  return data;
}
