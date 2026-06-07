"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getPayments, makePayment } from "@/lib/api/payment.service";

export function usePayments(params?: { page?: number; status?: string }) {
  return useQuery({
    queryKey: ["payments", params],
    queryFn: () => getPayments(params),
  });
}

export function useMakePayment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: makePayment,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["payments"] });
    },
  });
}
