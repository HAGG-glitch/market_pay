"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getLoans,
  getLoan,
  applyLoan,
  updateLoanStatus,
} from "@/lib/api/loan.service";

export function useLoans(params?: { page?: number; status?: string }) {
  return useQuery({
    queryKey: ["loans", params],
    queryFn: () => getLoans(params),
  });
}

export function useLoan(id: string) {
  return useQuery({
    queryKey: ["loan", id],
    queryFn: () => getLoan(id),
    enabled: !!id,
  });
}

export function useApplyLoan() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: applyLoan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["loans"] });
    },
  });
}

export function useUpdateLoanStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      status,
    }: {
      id: string;
      status: string;
    }) => updateLoanStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["loans"] });
    },
  });
}
