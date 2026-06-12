import apiClient from "./client";

export interface AuditLogEntry {
  id: string;
  actor_id: string;
  actor_role: string;
  action: string;
  resource: string;
  resource_id: string;
  old_state: string;
  new_state: string;
  created_at: string;
}

export async function getAuditLogs(params?: {
  actor_id?: string;
  resource?: string;
  since?: string;
}) {
  const { data } = await apiClient.get<{ data: AuditLogEntry[] }>("/audit-logs", {
    params,
  });
  return data.data ?? [];
}
