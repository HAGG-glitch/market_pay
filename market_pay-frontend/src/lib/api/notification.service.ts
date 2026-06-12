import apiClient from "./client";

export interface InAppNotification {
  id: string;
  event_type: string;
  title: string;
  body: string;
  is_read: boolean;
  created_at: string;
}

export async function getNotifications() {
  const { data } = await apiClient.get<{ data: InAppNotification[] }>(
    "/notifications"
  );
  return data.data ?? [];
}

export async function markNotificationRead(id: string) {
  await apiClient.put(`/notifications/${id}/read`);
}

export function subscribeNotifications(
  onEvent: (notification: InAppNotification) => void
) {
  const token = localStorage.getItem("marketpay_token");
  const mode = localStorage.getItem("marketpay_mode") || "demo";
  const base = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
  const url = `${base}/notifications/stream`;

  const es = new EventSource(url, { withCredentials: false });
  // EventSource cannot set headers — use query token if backend supports it.
  // Fallback: poll via getNotifications in hook.
  void token;
  void mode;

  es.onmessage = (ev) => {
    try {
      const payload = JSON.parse(ev.data);
      onEvent(payload as InAppNotification);
    } catch {
      /* ignore malformed */
    }
  };

  return () => es.close();
}
