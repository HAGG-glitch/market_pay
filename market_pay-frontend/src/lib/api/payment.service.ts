import apiClient from "./client";
import type { Payment, PaginatedResponse } from "@/types";

export async function getPayments(params?: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  const { data } = await apiClient.get<PaginatedResponse<Payment>>(
    "/payments",
    { params }
  );
  return data;
}

export async function makePayment(payload: {
  vendor_id: string;
  amount: number;
}) {
  const { data } = await apiClient.post<Payment>("/payments", payload);
  return data;
}
