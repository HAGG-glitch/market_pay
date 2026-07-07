"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getNotifications,
  markNotificationRead,
  type InAppNotification,
} from "@/lib/api/notification.service";

export function useNotifications() {
  return useQuery({
    queryKey: ["notifications"],
    queryFn: getNotifications,
    refetchInterval: 15_000,
  });
}

export function useMarkNotificationRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: markNotificationRead,
    onMutate: async (id: string) => {
      await qc.cancelQueries({ queryKey: ["notifications"] });
      const previous = qc.getQueryData<InAppNotification[]>(["notifications"]);
      qc.setQueryData<InAppNotification[]>(["notifications"], (old) =>
        (old ?? []).map((n) => (n.id === id ? { ...n, is_read: true } : n))
      );
      return { previous };
    },
    onError: (_err, _id, context) => {
      if (context?.previous) {
        qc.setQueryData(["notifications"], context.previous);
      }
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ["notifications"] }),
  });
}
