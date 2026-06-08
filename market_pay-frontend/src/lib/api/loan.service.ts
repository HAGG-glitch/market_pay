import apiClient from "./client";
import type { Loan, PaginatedResponse } from "@/types";

export async function getLoans(params?: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  const { data } = await apiClient.get<PaginatedResponse<Loan>>("/loans", {
    params,
  });
  return data;
}

export async function getLoan(id: string) {
  const { data } = await apiClient.get<Loan>(`/loans/${id}`);
  return data;
}

export async function applyLoan(payload: {
  amount: number;
  interest_rate: number;
  repayment_schedule: { due_date: string; amount: number }[];
}) {
  const { data } = await apiClient.post<Loan>("/loans", payload);
  return data;
}

export async function updateLoanStatus(id: string, status: string) {
  const { data } = await apiClient.patch<Loan>(`/loans/${id}/status`, {
    status,
  });
  return data;
}
